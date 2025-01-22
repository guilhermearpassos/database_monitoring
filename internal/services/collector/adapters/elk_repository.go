package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"log"
	"time"
)

type ELKRepository struct {
	client *elasticsearch.Client
}

func NewELKRepository(client *elasticsearch.Client) *ELKRepository {
	return &ELKRepository{client: client}
}

var _ domain.SampleRepository = (*ELKRepository)(nil)

func (r ELKRepository) StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error {

	err := r.storeSnapData(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("failed to store snapshot data: %w", err)
	}
	err = r.storeSamples(ctx, snapshot.Samples)
	if err != nil {
		return fmt.Errorf("store samples: %w", err)
	}
	return nil
}

func (r ELKRepository) storeSnapData(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error {
	jsonData, err := json.Marshal(snapshot.SnapInfo)
	if err != nil {
		return fmt.Errorf("marshal snapshot info: %w", err)
	}
	resp, err := r.client.Index("db_snapshots", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("index db_snapshots: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("index db_snapshots: %s", resp.String())
	}
	return nil
}

func (r ELKRepository) storeSamples(ctx context.Context, samples []*common_domain.QuerySample) error {
	var err error
	var indexer esutil.BulkIndexer
	indexer, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{Index: "db_samples", Client: r.client,
		NumWorkers:    5,
		FlushBytes:    1000,
		FlushInterval: 10 * time.Second,
		OnError: func(ctx context.Context, err error) {
			log.Printf("Error bulk indexing data: %w", err)
		},
	})
	if err != nil {
		return err
	}
	for _, sample := range samples {
		var jsonData []byte
		jsonData, err = json.Marshal(sample)
		if err != nil {
			err = fmt.Errorf("Error marshalling sample to JSON: %w", err)
			return err
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
			return err
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
	return nil
}
