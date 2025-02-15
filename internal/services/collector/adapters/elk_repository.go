package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"io"
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
var _ domain.QueryMetricsRepository = (*ELKRepository)(nil)

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

func (r ELKRepository) StoreQueryMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, timestamp time.Time) error {
	var err error
	var indexer esutil.BulkIndexer
	indexer, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{Index: "query_metrics", Client: r.client,
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
	for _, metric := range metrics {
		if metric.CollectionTime.IsZero() {
			metric.CollectionTime = timestamp
		}
		var jsonData []byte
		jsonData, err = json.Marshal(metric)
		if err != nil {
			err = fmt.Errorf("Error marshalling metric to JSON: %w", err)
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

func (r ELKRepository) StoreExecutionPlans(ctx context.Context, plans []*common_domain.ExecutionPlan) error {
	var err error
	var indexer esutil.BulkIndexer
	indexer, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{Index: "exec_plans", Client: r.client,
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
	for _, plan := range plans {
		var jsonData []byte
		jsonData, err = json.Marshal(plan)
		if err != nil {
			err = fmt.Errorf("Error marshalling plan to JSON: %w", err)
			return err
		}
		//s := string(plan.PlanHandle)
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

func (r ELKRepository) GetKnownPlanHandles(ctx context.Context, server *common_domain.ServerMeta) ([]string, error) {

	// Define the query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"Server.Host": server.Host,
						},
					},
					{
						"term": map[string]interface{}{
							"Server.Type": server.Type,
						},
					},
				},
			},
		},
		"_source": []string{"PlanHandle"}, // Keep only the PlanHandle field
	}
	//
	//// Convert the query to JSON
	//queryJSON, err := json.Marshal(query)
	//if err != nil {
	//	log.Fatalf("Error marshaling the query: %s", err)
	//}
	resp2, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("exec_plans"),
		r.client.Search.WithBody(esutil.NewJSONReader(&query)),
		r.client.Search.WithSize(10000),
		//r.client.Search.WithSort("Snapshot.ID, Session.SessionID"),
	)
	if err != nil {
		return nil, fmt.Errorf("getting exec plans: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp2.Body)
	if resp2.IsError() {
		if resp2.StatusCode == 404 {
			return nil, custom_errors.NotFoundErr{Message: fmt.Sprintf("getting exec plans: (code %d), %s: %s", resp2.StatusCode, resp2.Status(), resp2.String())}
		}
		return nil, fmt.Errorf("getting exec plans: (code %d), %s: %s", resp2.StatusCode, resp2.Status(), resp2.String())
	}

	var decodedResp SearchPlansResponse
	err = json.NewDecoder(resp2.Body).Decode(&decodedResp)
	if err != nil {
		print(resp2.String())
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	ret := make([]string, len(decodedResp.Hits.Hits))
	for i, h := range decodedResp.Hits.Hits {
		ret[i] = h.Source.PlanHandle
	}

	return ret, nil

}

type SearchPlansResponse struct {
	Took int64
	Hits struct {
		Total struct {
			Value int64
		}
		Hits []*SearchPlansHit
	}
}

type SearchPlansHit struct {
	Index   string                      `json:"_index"`
	ID      string                      `json:"_id"`
	Score   float64                     `json:"_score"`
	Ignored []string                    `json:"_ignored"`
	Source  struct{ PlanHandle string } `json:"_source"`
	Type    string                      `json:"_type"`
	Version int64                       `json:"_version,omitempty"`
}
