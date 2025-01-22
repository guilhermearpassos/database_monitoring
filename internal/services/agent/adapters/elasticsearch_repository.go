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

func (r ELKRepository) ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int) ([]common_domain.DataBaseSnapshot, int, error) {

	//ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	//defer cancel()

	snapshotInfos, ids, total, err3 := r.getSnapInfos(ctx, pageSize, pageNumber, "5", start, end)
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

func (r ELKRepository) getSnapInfos(ctx context.Context, pageSize int, pageNumber int, databaseID string, start time.Time, end time.Time) (map[string]common_domain.SnapInfo, []string, int, error) {
	from := (pageNumber - 1) * pageSize
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
		r.client.Search.WithSize(pageSize),
		r.client.Search.WithFrom(from),
		r.client.Search.WithTrackTotalHits(true),
		r.client.Search.WithContext(ctx),
		r.client.Search.WithBody(esutil.NewJSONReader(query)),
		r.client.Search.WithSort("timestamp"),
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
