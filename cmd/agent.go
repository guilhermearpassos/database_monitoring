package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/service"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"time"
)

var (
	configFileName string
	AgentCmd       = &cobra.Command{
		Use:     "agent",
		Short:   "run dbm agent",
		Long:    "run dbm agent",
		Aliases: []string{},
		Example: "dbm agent --config=local/agent.toml",
		RunE:    StartAgent,
	}
)

func init() {
	AgentCmd.Flags().StringVar(&configFileName, "config", "local/agent.toml", "--config=local/agent.toml")
}

func StartAgent(cmd *cobra.Command, args []string) error {
	var config service.AgentConfig
	// Check if file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		panic(fmt.Errorf("config file does not exist: %s", configFileName))
	}
	if _, err := toml.DecodeFile(configFileName, &config); err != nil {
		panic(fmt.Errorf("failed to parse config file: %s", err))
	}
	cc, err := grpc.NewClient(config.CollectorConfig.Url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := collectorv1.NewIngestionServiceClient(cc)
	for _, tgt := range config.TargetHosts {
		startTarget(tgt, client)
	}
	for {
		time.Sleep(5 * time.Second)
	}
	return nil
}
func startTarget(config service.TargetHostConfig, collectorClient collectorv1.IngestionServiceClient) {
	db, err := sqlx.Open(config.Driver, config.ConnString)
	if err != nil {
		panic(err)
	}
	knownHandles, err := collectorClient.GetKnownPlanHandles(context.Background(), &collectorv1.GetKnownPlanHandlesRequest{Server: &dbmv1.ServerMetadata{
		Host: config.Alias,
		Type: config.Driver,
	}})
	var knownHandlesSlice []string
	if err != nil {
		if grpcErr, ok := status.FromError(err); ok {
			if grpcErr.Code() == codes.NotFound {
				knownHandlesSlice = []string{}
			} else {
				panic(err)
			}

		} else {

			panic(err)
		}
	} else {

		knownHandlesSlice = knownHandles.Handles
	}

	dataReader := adapters.NewSQLServerDataReader(db, common_domain.ServerMeta{
		Host: config.Alias,
		Type: config.Driver,
	}, knownHandlesSlice)
	go collectSnapshots(dataReader, collectorClient)
	go collectQueryMetrics(dataReader, collectorClient, config.Alias)
}

func collectQueryMetrics(reader domain.QueryMetricsReader, client collectorv1.IngestionServiceClient, serverName string) {
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
			ServerId:  serverName,
			Timestamp: timestamppb.New(sampleTime),
			Metrics:   &collectorv1.DatabaseMetrics_QueryMetrics{QueryMetrics: &collectorv1.DatabaseMetrics_QueryMetricSample{QueryMetrics: protoMetrics}},
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
		planHandleStrings := make(map[string]struct{})
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
			for _, qs := range snapshot.Samples {
				if qs.PlanHandle == nil {
					continue
				}
				planHandleStrings[string(qs.PlanHandle)] = struct{}{}
			}

		}
		planHandles := make([][]byte, 0)
		for k := range planHandleStrings {
			planHandles = append(planHandles, []byte(k))
		}
		executionPlans, err := dataReader.GetPlanHandles(context.Background(), planHandles, true)
		if err != nil {
			panic(err)
		}
		protoPlans := make([]*dbmv1.ExecutionPlan, 0, len(executionPlans))
		for _, p := range executionPlans {
			protoPlan, err2 := converters.ExecutionPlanToProto(p)
			if err2 != nil {
				panic(err2)
			}
			protoPlans = append(protoPlans, protoPlan)
		}
		if len(protoPlans) != 0 {

			_, err = client.IngestExecutionPlans(context.Background(), &collectorv1.IngestExecutionPlansRequest{Plans: protoPlans})
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
