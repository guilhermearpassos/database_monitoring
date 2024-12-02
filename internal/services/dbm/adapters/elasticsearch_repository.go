package adapters

import (
	"bytes"
	"context"
	"database_monitoring/internal/services/dbm/domain"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/google/uuid"
	"log"
	"time"
)

type ELKRepository struct {
	client *elasticsearch.Client
}

func NewELKRepository(client *elasticsearch.Client) *ELKRepository {
	return &ELKRepository{client: client}
}

func (r ELKRepository) StoreSamples(ctx context.Context, samples []*domain.QuerySample) (err error) {
	//res, _ := r.client.Indices.Create("test")
	//_ = res.Body.Close()
	var indexer esutil.BulkIndexer
	indexer, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{Index: "db_samples", Client: r.client,
		NumWorkers:    5,
		FlushBytes:    1000,
		FlushInterval: 10 * time.Second,
		OnError: func(ctx context.Context, err error) {
			log.Printf("Error bulk indexing data: %v", err)
		},
	})
	if err != nil {
		return
	}
	for _, sample := range samples {
		var jsonData []byte
		jsonData, err = json.Marshal(sample)
		if err != nil {
			err = fmt.Errorf("Error marshalling sample to JSON: %v", err)
			return
		}
		err = indexer.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: uuid.NewString(),
			Body:       bytes.NewReader(jsonData),
			// OnSuccess is called for each successful operation

			// OnFailure is called for each failed operation
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					log.Printf("ERROR: %s", err)
				} else {
					log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
				}
			},
		})
		if err != nil {
			return
		}
	}
	err = indexer.Close(ctx)
	if err != nil {
		return err
	}
	stats := indexer.Stats()
	if stats.NumFailed > 0 {
		return fmt.Errorf("index encountered %d errors", stats.NumFailed)
	}
	time.Sleep(5 * time.Second)
	return
}
