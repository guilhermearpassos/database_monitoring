package main

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

var (
	AgentCmd = &cobra.Command{
		Use:     "agent",
		Short:   "run dbm agent",
		Long:    "run dbm agent",
		Aliases: []string{},
		Example: "dbm agent",
		RunE:    StartAgent,
	}
)

func StartAgent(cmd *cobra.Command, args []string) error {
	cc, err := grpc.NewClient("localhost:7080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := collectorv1.NewIngestionServiceClient(cc)
	db, err := sqlx.Open("mssql", "server=localhost;port=1433;database=SQL_EXECUTION_ROUTER;user id=sa;password=SqlServer2019!")
	if err != nil {
		panic(err)
	}
	dataReader := adapters.NewSQLServerDataReader(db)
	go collectSnapshots(dataReader, client)
	go collectQueryMetrics(dataReader, client)
	for {
		time.Sleep(5 * time.Second)
	}
	return nil
}

func collectQueryMetrics(reader domain.QueryMetricsReader, client collectorv1.IngestionServiceClient) {
	for {
		sampleTime := time.Now()
		metrics, err := reader.CollectMetrics(context.Background())
		if err != nil {
			panic(err)
		}
		protoMetrics := make([]*dbmv1.QueryMetric, len(metrics))

		for i, m := range metrics {
			protoMetrics[i], err = converters.QueryMetricToProto(m)
			if err != nil {
				panic(err)
			}
		}
		_, err = client.IngestMetrics(context.Background(), &collectorv1.DatabaseMetrics{
			DatabaseId: "localhost",
			Timestamp:  timestamppb.New(sampleTime),
			Metrics:    &collectorv1.DatabaseMetrics_QueryMetrics{QueryMetrics: &collectorv1.DatabaseMetrics_QueryMetricSample{QueryMetrics: protoMetrics}},
		})
		if err != nil {
			panic(err)
		}

		select {
		case <-time.After(time.Until(sampleTime.Add(1 * time.Minute))):
			break

		}
	}
}

func collectSnapshots(dataReader domain.SamplesReader, client collectorv1.IngestionServiceClient) {
	for {
		var snapshots []*common_domain.DataBaseSnapshot
		snapshots, err := dataReader.TakeSnapshot(context.Background())
		if err != nil {
			panic(err)
		}
		for _, snapshot := range snapshots {
			if snapshot == nil {
				continue
			}
			_, err = client.IngestSnapshot(context.Background(), &collectorv1.IngestSnapshotRequest{
				Snapshot: converters.DatabaseSnapshotToProto(snapshot),
			})
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
