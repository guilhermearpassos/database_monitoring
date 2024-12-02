//go:debug x509negativeserial=1
package main

import (
	"context"
	"crypto/tls"
	"database_monitoring/internal/services/dbm/adapters"
	"database_monitoring/internal/services/dbm/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jmoiron/sqlx"
	"net/http"
	"time"
)

func main() {
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
		querySamples := make([]*domain.QuerySample, 0)
		for _, snapshot := range snapshots {
			querySamples = append(querySamples, snapshot.Samples...)
		}
		err = elk.StoreSamples(context.Background(), querySamples)
		if err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Second)
	}
	//c := &cobra.Command{
	//	Use: "er",
	//	RunE: func(cmd *cobra.Command, args []string) error {
	//		return cmd.Usage()
	//	},
	//}
	//c.AddCommand(MigrateCmd)
	//c.AddCommand(GrpcCmd)
	//err := c.Execute()
	//if err != nil {
	//	panic(err)
	//}
}
