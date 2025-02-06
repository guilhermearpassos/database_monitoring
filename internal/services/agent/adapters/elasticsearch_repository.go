package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	io "io"
	"strings"
	"time"
)

type ELKRepository struct {
	client *elasticsearch.Client
}

func NewELKRepository(client *elasticsearch.Client) *ELKRepository {
	return &ELKRepository{client: client}
}

func (r ELKRepository) ListServers(ctx context.Context, start time.Time, end time.Time) ([]domain.ServerSummary, error) {

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"range": map[string]interface{}{
						"timestamp": map[string]interface{}{
							"gte": start,
							"lte": end,
						},
					},
				},
			},
		},
	}
	resp, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("db_snapshots"),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithContext(ctx),
		r.client.Search.WithBody(esutil.NewJSONReader(query)),
		r.client.Search.WithStats("| STATS last_snap = MAX(timestamp)  by server.Type, server.Host \n"),
		r.client.Search.WithSort("timestamp"),
	)
	if err != nil {
		return nil, fmt.Errorf("searching serverMeta: %w", err)
	}
	defer func(Body *io.ReadCloser) {
		if Body == nil {
			return
		}
		_ = (*Body).Close()
	}(&resp.Body)
	if resp.IsError() {
		return nil, fmt.Errorf("getting serverMeta: (code %d), %s:  %s", resp.StatusCode, resp.Status(), resp.String())
	}
	var decodedResp SearchServersResponse
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	//timestamps := make([]time.Time, len(decodedResp.Hits.Hits))
	//return ret, nil
	panic("unimplemented")
}

func (r ELKRepository) ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error) {

	//ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	//defer cancel()

	snapshotInfos, ids, total, err3 := r.getSnapInfos(ctx, pageSize, pageNumber, start, end, serverID)
	if err3 != nil {
		return nil, 0, err3
	}
	if total == 0 {
		return []common_domain.DataBaseSnapshot{}, 0, nil
	}
	if len(ids) == 0 {
		return []common_domain.DataBaseSnapshot{}, 0, nil
	}
	samplesBySnap, err2 := r.getSnapSamples(ctx, ids)
	if err2 != nil {
		return nil, 0, err2
	}
	snapshots := make([]common_domain.DataBaseSnapshot, 0)
	for id, snapInfo := range snapshotInfos {
		snap := common_domain.DataBaseSnapshot{
			Samples:  samplesBySnap[id],
			SnapInfo: snapInfo,
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, total, nil
}

func (r ELKRepository) getSnapInfos(ctx context.Context, pageSize int, pageNumber int, start time.Time, end time.Time, serverID string) (map[string]common_domain.SnapInfo, []string, int, error) {
	from := (pageNumber - 1) * pageSize
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"server.Host.keyword": serverID,
						},
					},
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": start,
								"lte": end,
							},
						},
					},
				},
			},
		},
	}
	resp, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("db_snapshots"),
		r.client.Search.WithSize(pageSize),
		r.client.Search.WithFrom(from),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithContext(ctx),
		r.client.Search.WithBody(esutil.NewJSONReader(query)),
		r.client.Search.WithSort("timestamp:desc"),
	)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("getting snapshots: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, nil, 0, fmt.Errorf("getting snapshots: (code %d), %s:  %s", resp.StatusCode, resp.Status(), resp.String())
	}
	var decodedResp SearchSnapResponse
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("decoding response body: %w", err)
	}
	snapshotInfos := make(map[string]common_domain.SnapInfo, len(decodedResp.Hits.Hits))
	total := int(decodedResp.Hits.Total.Value)
	ids := make([]string, 0)
	for _, si := range decodedResp.Hits.Hits {
		snapshotInfos[si.Source.ID] = si.Source
		ids = append(ids, si.Source.ID)
	}
	return snapshotInfos, ids, total, nil
}

func (r ELKRepository) getSnapSamples(ctx context.Context, ids []string) (map[string][]*common_domain.QuerySample, error) {

	queryString := "Snapshot.ID in (" + strings.Join(ids, ",") + ")"

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": queryString,
			},
		},
	}
	resp2, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("db_samples"),
		r.client.Search.WithBody(esutil.NewJSONReader(&query)),
		r.client.Search.WithSize(10000),
		//r.client.Search.WithSort("Snapshot.ID, Session.SessionID"),
	)
	if err != nil {
		return nil, fmt.Errorf("getting samples: %w", err)
	}
	defer resp2.Body.Close()
	if resp2.IsError() {
		return nil, fmt.Errorf("getting samples: (code %d), %s: %s", resp2.StatusCode, resp2.Status(), resp2.String())
	}
	var decodedResp2 SearchSamplesResponse
	err = json.NewDecoder(resp2.Body).Decode(&decodedResp2)
	if err != nil {
		print(resp2.String())
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	samplesBySnap := make(map[string][]*common_domain.QuerySample, len(decodedResp2.Hits.Hits))
	for _, h := range decodedResp2.Hits.Hits {
		sample := h.Source
		if _, ok := samplesBySnap[sample.Snapshot.ID]; ok {
			samplesBySnap[sample.Snapshot.ID] = append(samplesBySnap[sample.Snapshot.ID], &sample)
		} else {
			samplesBySnap[sample.Snapshot.ID] = []*common_domain.QuerySample{&sample}
		}
	}

	return samplesBySnap, nil
}

type SearchSnapResponse struct {
	Took int64
	Hits struct {
		Total struct {
			Value int64
		}
		Hits []*SearchDBSnapHit
	}
}

type SearchDBSnapHit struct {
	Index   string                 `json:"_index"`
	ID      string                 `json:"_id"`
	Score   float64                `json:"_score"`
	Ignored []string               `json:"_ignored"`
	Source  common_domain.SnapInfo `json:"_source"`
	Type    string                 `json:"_type"`
	Version int64                  `json:"_version,omitempty"`
}

type SearchSamplesResponse struct {
	Took int64
	Hits struct {
		Total struct {
			Value int64
		}
		Hits []*SearchSamplesHit
	}
}

type SearchSamplesHit struct {
	Index   string                    `json:"_index"`
	ID      string                    `json:"_id"`
	Score   float64                   `json:"_score"`
	Ignored []string                  `json:"_ignored"`
	Source  common_domain.QuerySample `json:"_source"`
	Type    string                    `json:"_type"`
	Version int64                     `json:"_version,omitempty"`
}

type SearchServersResponse struct {
	Took int64
	Hits struct {
		Total struct {
			Value int64
		}
		Hits []*SearchServerHit
	}
}
type SearchServerHit struct {
	Index   string                          `json:"_index"`
	ID      string                          `json:"_id"`
	Score   float64                         `json:"_score"`
	Ignored []string                        `json:"_ignored"`
	Source  struct{ Server ServerLastSnap } `json:"_source"`
	Type    string                          `json:"_type"`
	Version int64                           `json:"_version,omitempty"`
}
type ServerLastSnap struct {
	Server struct {
		Host string `json:"host"`
		Type string `json:"type"`
	}
	LastSnap time.Time `json:"last_snap"`
}

func (r ELKRepository) GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"id.keyword": id,
						},
					},
				},
			},
		},
	}
	resp, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("db_snapshots"),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithContext(ctx),
		r.client.Search.WithBody(esutil.NewJSONReader(query)),
	)
	if err != nil {
		return common_domain.DataBaseSnapshot{}, fmt.Errorf("getting snapshot: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return common_domain.DataBaseSnapshot{}, fmt.Errorf("getting snapshot: (code %d), %s:  %s", resp.StatusCode, resp.Status(), resp.String())
	}
	var decodedResp SearchSnapResponse
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return common_domain.DataBaseSnapshot{}, fmt.Errorf("decoding response body: %w", err)
	}
	if len(decodedResp.Hits.Hits) == 0 {
		return common_domain.DataBaseSnapshot{}, fmt.Errorf("snapshot %s not found", id)
	}
	snapInfo := decodedResp.Hits.Hits[0].Source
	snapSamples, err := r.getSnapSamples(ctx, []string{id})
	if err != nil {
		return common_domain.DataBaseSnapshot{}, err
	}
	snap := common_domain.DataBaseSnapshot{
		Samples:  snapSamples[id],
		SnapInfo: snapInfo,
	}

	return snap, nil
}

func (r ELKRepository) GetQueryMetrics(ctx context.Context, start time.Time, end time.Time) ([]*common_domain.QueryMetric, error) {

	query := fmt.Sprintf(`{
  "aggs": {
    "by_queryhash": {
      "aggs": {
        "executionCount": {
          "sum": {
            "field": "Counters.executionCount"
          }
        },
        "totalClrTime": {
          "sum": {
            "field": "Counters.totalClrTime"
          }
        },
        "totalColumnstoreSegmentReads": {
          "sum": {
            "field": "Counters.totalColumnstoreSegmentReads"
          }
        },
        "totalColumnstoreSegmentSkips": {
          "sum": {
            "field": "Counters.totalColumnstoreSegmentSkips"
          }
        },
        "totalDop": {
          "sum": {
            "field": "Counters.totalDop"
          }
        },
        "totalElapsedTime": {
          "sum": {
            "field": "Counters.totalElapsedTime"
          }
        },
        "totalGrantKb": {
          "sum": {
            "field": "Counters.totalGrantKb"
          }
        },
        "totalIdealGrantKb": {
          "sum": {
            "field": "Counters.totalIdealGrantKb"
          }
        },
        "totalLogicalReads": {
          "sum": {
            "field": "Counters.totalLogicalReads"
          }
        },
        "totalLogicalWrites": {
          "sum": {
            "field": "Counters.totalLogicalWrites"
          }
        },
        "totalPhysicalReads": {
          "sum": {
            "field": "Counters.totalPhysicalReads"
          }
        },
        "totalReservedThreads": {
          "sum": {
            "field": "Counters.totalReservedThreads"
          }
        },
        "totalRows": {
          "sum": {
            "field": "Counters.totalRows"
          }
        },
        "totalSpills": {
          "sum": {
            "field": "Counters.totalSpills"
          }
        },
        "totalUsedGrantKb": {
          "sum": {
            "field": "Counters.totalUsedGrantKb"
          }
        },
        "totalUsedThreads": {
          "sum": {
            "field": "Counters.totalUsedThreads"
          }
        },
        "totalWorkerTime": {
          "sum": {
            "field": "Counters.totalWorkerTime"
          }
        },
        "last_exec": {
          "max": {
            "field": "LastExecutionTime"
          }
        }
      },
      "terms": {
        "field": "QueryHash.keyword"
      }
    }
  },
  "query": {
    "bool": {
      "must": {
        "range": {
          "CollectionTime": {
            "gte": "%s",
            "lte": "%s"
          }
        }
      }
    }
  }
}`, start.Format("2006-01-02T15:04:05"), end.Format("2006-01-02T15:04:05"))
	jsonBody, err := json.Marshal(query)
	fmt.Println(string(jsonBody))
	//esutil.NewJSONReader(query).Read(jsonBody)
	resp, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("query_metrics"),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, fmt.Errorf("searching serverMeta: %w", err)
	}
	defer func(Body *io.ReadCloser) {
		if Body == nil {
			return
		}
		_ = (*Body).Close()
	}(&resp.Body)
	if resp.IsError() {
		return nil, fmt.Errorf("getting serverMeta: (code %d), %s:  %s", resp.StatusCode, resp.Status(), resp.String())
	}
	//rawMessage := json.RawMessage{}
	//json.NewDecoder(resp.Body).Decode(&rawMessage)
	var decodedResp QueryMetricsResponse
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	//timestamps := make([]time.Time, len(decodedResp.Hits.Hits))
	//return ret, nil
	ret := make([]*common_domain.QueryMetric, len(decodedResp.Aggregations.ByQueryhash.Buckets))
	for i, agg := range decodedResp.Aggregations.ByQueryhash.Buckets {
		lasExec := time.Unix(0, int64(agg.LastExecutionTime.Value)*1000000)
		//if err != nil {
		//	return nil, fmt.Errorf("parsing last execution time: %w", err)
		//}

		executionCount := agg.ExecutionCount["value"]
		counters := map[string]int64{
			"executionCount":               int64(executionCount),
			"totalWorkerTime":              int64(agg.TotalWorkerTime["value"]),
			"totalPhysicalReads":           int64(agg.TotalPhysicalReads["value"]),
			"totalLogicalWrites":           int64(agg.TotalLogicalWrites["value"]),
			"totalLogicalReads":            int64(agg.TotalLogicalReads["value"]),
			"totalClrTime":                 int64(agg.TotalClrTime["value"]),
			"totalElapsedTime":             int64(agg.TotalElapsedTime["value"]),
			"totalRows":                    int64(agg.TotalRows["value"]),
			"totalDop":                     int64(agg.TotalDop["value"]),
			"totalGrantKb":                 int64(agg.TotalGrantKb["value"]),
			"totalUsedGrantKb":             int64(agg.TotalUsedGrantKb["value"]),
			"totalIdealGrantKb":            int64(agg.TotalIdealGrantKb["value"]),
			"totalReservedThreads":         int64(agg.TotalReservedThreads["value"]),
			"totalUsedThreads":             int64(agg.TotalUsedThreads["value"]),
			"totalColumnstoreSegmentReads": int64(agg.TotalColumnstoreSegmentReads["value"]),
			"totalColumnstoreSegmentSkips": int64(agg.TotalColumnstoreSegmentSkips["value"]),
			"totalSpills":                  int64(agg.TotalSpills["value"]),
		}
		rates := map[string]float64{
			"avgWorkerTime":              agg.TotalWorkerTime["value"] / executionCount,
			"avgPhysicalReads":           agg.TotalPhysicalReads["value"] / executionCount,
			"avgLogicalWrites":           agg.TotalLogicalWrites["value"] / executionCount,
			"avgLogicalReads":            agg.TotalLogicalReads["value"] / executionCount,
			"avgClrTime":                 agg.TotalClrTime["value"] / executionCount,
			"avgElapsedTime":             agg.TotalElapsedTime["value"] / executionCount,
			"avgRows":                    agg.TotalRows["value"] / executionCount,
			"avgDop":                     agg.TotalDop["value"] / executionCount,
			"avgGrantKb":                 agg.TotalGrantKb["value"] / executionCount,
			"avgUsedGrantKb":             agg.TotalUsedGrantKb["value"] / executionCount,
			"avgIdealGrantKb":            agg.TotalIdealGrantKb["value"] / executionCount,
			"avgReservedThreads":         agg.TotalReservedThreads["value"] / executionCount,
			"avgUsedThreads":             agg.TotalUsedThreads["value"] / executionCount,
			"avgColumnstoreSegmentReads": agg.TotalColumnstoreSegmentReads["value"] / executionCount,
			"avgColumnstoreSegmentSkips": agg.TotalColumnstoreSegmentSkips["value"] / executionCount,
			"avgSpills":                  agg.TotalSpills["value"] / executionCount,
		}
		ret[i] = &common_domain.QueryMetric{
			QueryHash:         agg.QueryHash,
			Text:              "",
			Database:          common_domain.DataBaseMetadata{},
			LastExecutionTime: lasExec,
			LastElapsedTime:   0,
			Counters:          counters,
			Rates:             rates,
			CollectionTime:    time.Time{},
		}
	}
	return ret, nil
}

type QueryMetricsResponse struct {
	Took         int64
	Aggregations struct {
		ByQueryhash struct {
			Buckets []*QueryMetricsHit
		} `json:"by_queryhash"`
	}
}
type QueryMetricsHit struct {
	ExecutionCount               map[string]float64 `json:"executionCount"`
	TotalWorkerTime              map[string]float64 `json:"totalWorkerTime"`
	TotalPhysicalReads           map[string]float64 `json:"totalPhysicalReads"`
	TotalLogicalWrites           map[string]float64 `json:"totalLogicalWrites"`
	TotalLogicalReads            map[string]float64 `json:"totalLogicalReads"`
	TotalClrTime                 map[string]float64 `json:"totalClrTime"`
	TotalElapsedTime             map[string]float64 `json:"totalElapsedTime"`
	TotalRows                    map[string]float64 `json:"totalRows"`
	TotalDop                     map[string]float64 `json:"totalDop"`
	TotalGrantKb                 map[string]float64 `json:"totalGrantKb"`
	TotalUsedGrantKb             map[string]float64 `json:"totalUsedGrantKb"`
	TotalIdealGrantKb            map[string]float64 `json:"totalIdealGrantKb"`
	TotalReservedThreads         map[string]float64 `json:"totalReservedThreads"`
	TotalUsedThreads             map[string]float64 `json:"totalUsedThreads"`
	TotalColumnstoreSegmentReads map[string]float64 `json:"totalColumnstoreSegmentReads"`
	TotalColumnstoreSegmentSkips map[string]float64 `json:"totalColumnstoreSegmentSkips"`
	TotalSpills                  map[string]float64 `json:"totalSpills"`
	LastExecutionTime            struct {
		Value         float64
		ValueAsString string
	} `json:"last_exec"`
	QueryHash []byte `json:"key"`
}
