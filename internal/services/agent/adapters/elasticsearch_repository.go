package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"io"
	"strings"
	"time"
)

type ELKRepository struct {
	client *elasticsearch.Client
	tracer trace.Tracer
}

func NewELKRepository(client *elasticsearch.Client) *ELKRepository {
	return &ELKRepository{client: client, tracer: otel.Tracer("ELKRepository")}
}

func (r ELKRepository) ListServers(ctx context.Context, start time.Time, end time.Time) ([]domain.ServerSummary, error) {

	query := fmt.Sprintf(`{
		"query":  {
			"bool":  {
				"must":  {
					"range":  {
						"timestamp":  {
							"gte": "%s",
							"lte": "%s"
						}
					}
				}
			}
		},
		"aggs":{
			"unique_ids": {
				"terms": {
					"field": "server.Host.keyword"
				}
			}
		}
	}`, start.Format("2006-01-02T15:04:05"), end.Format("2006-01-02T15:04:05"))

	jsonBody, err := json.Marshal(query)
	fmt.Println(string(jsonBody))
	resp, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("db_snapshots"),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithContext(ctx),
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
	//fmt.Println(string(rawMessage))
	var decodedResp SearchServersResponse
	err = json.NewDecoder(resp.Body).Decode(&decodedResp)
	if err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	//timestamps := make([]time.Time, len(decodedResp.Hits.Hits))
	ret := make([]domain.ServerSummary, len(decodedResp.Aggregations.UniqueIds.Buckets))
	for i, h := range decodedResp.Aggregations.UniqueIds.Buckets {
		ret[i] = domain.ServerSummary{
			Name:             h.Key,
			Type:             "mssql",
			Connections:      0,
			RequestRate:      0,
			ConnsByWaitGroup: nil,
		}
	}
	return ret, nil
	//panic("unimplemented")
}

func (r ELKRepository) ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error) {
	ctx, span := r.tracer.Start(ctx, "ListSnapshots")
	defer span.End()
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
	ctx, span := r.tracer.Start(ctx, "getSnapInfos")
	defer span.End()
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
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
	ctx, span := r.tracer.Start(ctx, "getSnapSamples")
	defer span.End()
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp2.Body)
	if resp2.IsError() {
		return nil, fmt.Errorf("getting samples: (code %d), %s: %s", resp2.StatusCode, resp2.Status(), resp2.String())
	}
	var decodedResp2 SearchSamplesResponse
	ctx, span2 := r.tracer.Start(ctx, "getSnapSamples-decode")
	defer span2.End()
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
	Aggregations struct {
		UniqueIds struct {
			Buckets []struct {
				Key string `json:"key"`
			} `json:"buckets"`
		} `json:"unique_ids"`
	} `json:"aggregations"`
}
type SearchServerHit struct {
	Index   string         `json:"_index"`
	ID      string         `json:"_id"`
	Score   float64        `json:"_score"`
	Ignored []string       `json:"_ignored"`
	Source  ServerLastSnap `json:"_source"`
	Type    string         `json:"_type"`
	Version int64          `json:"_version,omitempty"`
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
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
        "sample": {
          "top_hits": {
            "size": 1
          }
        },
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
		lasExec := agg.LastExecutionTime()
		sample := agg.Sample()
		if sample == nil {
			continue
		}
		executionCount := agg.ExecutionCount()
		counters := map[string]int64{
			"executionCount":               executionCount,
			"totalWorkerTime":              agg.TotalWorkerTime(),
			"totalPhysicalReads":           agg.TotalPhysicalReads(),
			"totalLogicalWrites":           agg.TotalLogicalWrites(),
			"totalLogicalReads":            agg.TotalLogicalReads(),
			"totalClrTime":                 agg.TotalClrTime(),
			"totalElapsedTime":             agg.TotalElapsedTime(),
			"totalRows":                    agg.TotalRows(),
			"totalDop":                     agg.TotalDop(),
			"totalGrantKb":                 agg.TotalGrantKb(),
			"totalUsedGrantKb":             agg.TotalUsedGrantKb(),
			"totalIdealGrantKb":            agg.TotalIdealGrantKb(),
			"totalReservedThreads":         agg.TotalReservedThreads(),
			"totalUsedThreads":             agg.TotalUsedThreads(),
			"totalColumnstoreSegmentReads": agg.TotalColumnstoreSegmentReads(),
			"totalColumnstoreSegmentSkips": agg.TotalColumnstoreSegmentSkips(),
			"totalSpills":                  agg.TotalSpills(),
		}
		rates := map[string]float64{
			"avgWorkerTime":              float64(agg.TotalWorkerTime()) / float64(executionCount),
			"avgPhysicalReads":           float64(agg.TotalPhysicalReads()) / float64(executionCount),
			"avgLogicalWrites":           float64(agg.TotalLogicalWrites()) / float64(executionCount),
			"avgLogicalReads":            float64(agg.TotalLogicalReads()) / float64(executionCount),
			"avgClrTime":                 float64(agg.TotalClrTime()) / float64(executionCount),
			"avgElapsedTime":             float64(agg.TotalElapsedTime()) / float64(executionCount),
			"avgRows":                    float64(agg.TotalRows()) / float64(executionCount),
			"avgDop":                     float64(agg.TotalDop()) / float64(executionCount),
			"avgGrantKb":                 float64(agg.TotalGrantKb()) / float64(executionCount),
			"avgUsedGrantKb":             float64(agg.TotalUsedGrantKb()) / float64(executionCount),
			"avgIdealGrantKb":            float64(agg.TotalIdealGrantKb()) / float64(executionCount),
			"avgReservedThreads":         float64(agg.TotalReservedThreads()) / float64(executionCount),
			"avgUsedThreads":             float64(agg.TotalUsedThreads()) / float64(executionCount),
			"avgColumnstoreSegmentReads": float64(agg.TotalColumnstoreSegmentReads()) / float64(executionCount),
			"avgColumnstoreSegmentSkips": float64(agg.TotalColumnstoreSegmentSkips()) / float64(executionCount),
			"avgSpills":                  float64(agg.TotalSpills()) / float64(executionCount),
		}
		ret[i] = &common_domain.QueryMetric{
			QueryHash:         agg.QueryHash,
			Text:              sample.Text,
			Database:          sample.Database,
			LastExecutionTime: lasExec,
			LastElapsedTime:   0,
			Counters:          counters,
			Rates:             rates,
			CollectionTime:    time.Time{},
		}
	}
	return ret, nil
}

func (r ELKRepository) GetExecutionPlan(ctx context.Context, planHandle []byte, server *common_domain.ServerMeta) (*common_domain.ExecutionPlan, error) {

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
							"PlanHandle.keyword": base64.StdEncoding.EncodeToString(planHandle),
						},
					},
					//{
					//	"term": map[string]interface{}{
					//		"Server.Type": "mssql",
					//	},
					//},
				},
			},
		},
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

	var decodedResp struct {
		Took int64
		Hits struct {
			Total struct {
				Value int64
			}
			Hits []*struct {
				Index   string                       `json:"_index"`
				ID      string                       `json:"_id"`
				Score   float64                      `json:"_score"`
				Ignored []string                     `json:"_ignored"`
				Source  *common_domain.ExecutionPlan `json:"_source"`
				Type    string                       `json:"_type"`
				Version int64                        `json:"_version,omitempty"`
			}
		}
	}

	err = json.NewDecoder(resp2.Body).Decode(&decodedResp)
	if err != nil {
		print(resp2.String())
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	ret := make([]*common_domain.ExecutionPlan, len(decodedResp.Hits.Hits))
	for i, h := range decodedResp.Hits.Hits {
		ret[i] = h.Source
	}
	if len(ret) == 0 {
		return nil, custom_errors.NotFoundErr{Message: fmt.Sprintf("no plan found for handle %s", base64.StdEncoding.EncodeToString(planHandle))}
	}

	return ret[0], nil

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
	ExecutionCountMap               map[string]float64 `json:"executionCount"`
	TotalWorkerTimeMap              map[string]float64 `json:"totalWorkerTime"`
	TotalPhysicalReadsMap           map[string]float64 `json:"totalPhysicalReads"`
	TotalLogicalWritesMap           map[string]float64 `json:"totalLogicalWrites"`
	TotalLogicalReadsMap            map[string]float64 `json:"totalLogicalReads"`
	TotalClrTimeMap                 map[string]float64 `json:"totalClrTime"`
	TotalElapsedTimeMap             map[string]float64 `json:"totalElapsedTime"`
	TotalRowsMap                    map[string]float64 `json:"totalRows"`
	TotalDopMap                     map[string]float64 `json:"totalDop"`
	TotalGrantKbMap                 map[string]float64 `json:"totalGrantKb"`
	TotalUsedGrantKbMap             map[string]float64 `json:"totalUsedGrantKb"`
	TotalIdealGrantKbMap            map[string]float64 `json:"totalIdealGrantKb"`
	TotalReservedThreadsMap         map[string]float64 `json:"totalReservedThreads"`
	TotalUsedThreadsMap             map[string]float64 `json:"totalUsedThreads"`
	TotalColumnstoreSegmentReadsMap map[string]float64 `json:"totalColumnstoreSegmentReads"`
	TotalColumnstoreSegmentSkipsMap map[string]float64 `json:"totalColumnstoreSegmentSkips"`
	TotalSpillsMap                  map[string]float64 `json:"totalSpills"`
	LastExecutionTimeMap            struct {
		Value         float64
		ValueAsString string
	} `json:"last_exec"`
	SampleMap struct {
		Hits struct {
			Hits []struct {
				Source *common_domain.QueryMetric `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	} `json:"sample"`
	QueryHash []byte `json:"key"`
}

func (q QueryMetricsHit) ExecutionCount() int64 {
	return int64(q.ExecutionCountMap["value"])
}
func (q QueryMetricsHit) TotalWorkerTime() int64 {
	return int64(q.TotalWorkerTimeMap["value"])
}
func (q QueryMetricsHit) TotalPhysicalReads() int64 {
	return int64(q.TotalPhysicalReadsMap["value"])
}
func (q QueryMetricsHit) TotalLogicalWrites() int64 {
	return int64(q.TotalLogicalWritesMap["value"])
}
func (q QueryMetricsHit) TotalLogicalReads() int64 {
	return int64(q.TotalLogicalReadsMap["value"])
}
func (q QueryMetricsHit) TotalClrTime() int64 {
	return int64(q.TotalClrTimeMap["value"])
}
func (q QueryMetricsHit) TotalElapsedTime() int64 {
	return int64(q.TotalElapsedTimeMap["value"])
}
func (q QueryMetricsHit) TotalRows() int64 {
	return int64(q.TotalRowsMap["value"])
}
func (q QueryMetricsHit) TotalDop() int64 {
	return int64(q.TotalDopMap["value"])
}
func (q QueryMetricsHit) TotalGrantKb() int64 {
	return int64(q.TotalGrantKbMap["value"])
}
func (q QueryMetricsHit) TotalUsedGrantKb() int64 {
	return int64(q.TotalUsedGrantKbMap["value"])
}
func (q QueryMetricsHit) TotalIdealGrantKb() int64 {
	return int64(q.TotalIdealGrantKbMap["value"])
}
func (q QueryMetricsHit) TotalReservedThreads() int64 {
	return int64(q.TotalReservedThreadsMap["value"])
}
func (q QueryMetricsHit) TotalUsedThreads() int64 {
	return int64(q.TotalUsedThreadsMap["value"])
}
func (q QueryMetricsHit) TotalColumnstoreSegmentReads() int64 {
	return int64(q.TotalColumnstoreSegmentReadsMap["value"])
}
func (q QueryMetricsHit) TotalColumnstoreSegmentSkips() int64 {
	return int64(q.TotalColumnstoreSegmentSkipsMap["value"])
}
func (q QueryMetricsHit) TotalSpills() int64 {
	return int64(q.TotalSpillsMap["value"])
}
func (q QueryMetricsHit) LastExecutionTime() time.Time {
	return time.Unix(0, int64(q.LastExecutionTimeMap.Value)*1000000)
}

func (q QueryMetricsHit) Sample() *common_domain.QueryMetric {
	if len(q.SampleMap.Hits.Hits) == 0 {
		return nil
	}
	return q.SampleMap.Hits.Hits[0].Source
}
