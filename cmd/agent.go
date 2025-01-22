package main

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	for {
		var snapshots []*common_domain.DataBaseSnapshot
		snapshots, err = dataReader.TakeSnapshot(context.Background())
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
