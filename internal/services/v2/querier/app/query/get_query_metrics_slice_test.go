package query

import (
	"context"
	"testing"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/stretchr/testify/require"
)

type stubQueryMetricsRepo struct {
	slice []*common_domain.QueryMetric
}

func (s stubQueryMetricsRepo) ListQueryMetrics(context.Context, time.Time, time.Time, string) ([]*common_domain.QueryMetric, error) {
	return nil, nil
}

func (s stubQueryMetricsRepo) GetQueryMetricsSlice(context.Context, time.Time, time.Time, string, string) ([]*common_domain.QueryMetric, error) {
	return s.slice, nil
}

func TestGetQueryMetricsSliceHandler_AggregatesBucketsCorrectly(t *testing.T) {
	base := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	repo := stubQueryMetricsRepo{
		slice: []*common_domain.QueryMetric{
			{
				QueryHash:         "hash-1",
				CollectionTime:    base.Add(10 * time.Second),
				LastExecutionTime: base.Add(10 * time.Second),
				Counters:          map[string]int64{"executionCount": 2, "totalWorkerTime": 1000},
			},
			{
				QueryHash:         "hash-1",
				CollectionTime:    base.Add(40 * time.Second),
				LastExecutionTime: base.Add(40 * time.Second),
				Counters:          map[string]int64{"executionCount": 5, "totalWorkerTime": 3500},
			},
			{
				QueryHash:         "hash-1",
				CollectionTime:    base.Add(2*time.Minute + 5*time.Second),
				LastExecutionTime: base.Add(2*time.Minute + 5*time.Second),
				Counters:          map[string]int64{"executionCount": 3, "totalWorkerTime": 900},
			},
		},
	}

	h := NewGetQueryMetricsSliceHandler(repo)
	res, err := h.Handle(context.Background(), base, base.Add(5*time.Minute), "localhost", "hash-1", time.Minute)
	require.NoError(t, err)
	require.Len(t, res, 2)

	// Bucket 0 merges 2+5 = 7 (not 2+2 = 4!) and 1000+3500 = 4500
	require.Equal(t, int64(7), res[0].Counters["executionCount"])
	require.Equal(t, int64(4500), res[0].Counters["totalWorkerTime"])
	require.Equal(t, base, res[0].CollectionTime)

	// Bucket 2 has 3 and 900
	require.Equal(t, int64(3), res[1].Counters["executionCount"])
	require.Equal(t, int64(900), res[1].Counters["totalWorkerTime"])
	require.Equal(t, base.Add(2*time.Minute), res[1].CollectionTime)
}
