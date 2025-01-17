package main

import (
	"context"
	"crypto/tls"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/domain"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"net/http"
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
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     []string{"https://localhost:9200"},
		Username:      "elastic",
		Password:      "changeme",
		EnableMetrics: false,
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	})
	if err != nil {
		panic(err)
	}
	elk := adapters.NewELKRepository(client)
	db, err := sqlx.Open("mssql", "server=localhost;port=1433;database=SQL_EXECUTION_ROUTER;user id=sa;password=SqlServer2019!")
	if err != nil {
		panic(err)
	}
	dataReader := adapters.NewSQLServerDataReader(db)
	for {
		var snapshots []*domain.DataBaseSnapshot
		snapshots, err = dataReader.TakeSnapshot(context.Background())
		if err != nil {
			panic(err)
		}
		for _, snapshot := range snapshots {
			if snapshot == nil {
				continue
			}
			err = elk.StoreSnapshot(context.Background(), *snapshot)
			if err != nil {
				panic(err)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
