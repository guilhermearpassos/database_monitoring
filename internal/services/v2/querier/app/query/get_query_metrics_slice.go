package query

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
)

type GetQueryMetricsSliceHandler struct {
	repo domain.QueryMetricsRepository
}

func NewGetQueryMetricsSliceHandler(repo domain.QueryMetricsRepository) GetQueryMetricsSliceHandler {
	return GetQueryMetricsSliceHandler{repo: repo}
}

func (h GetQueryMetricsSliceHandler) Handle(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string, interval time.Duration) ([]*common_domain.QueryMetric, error) {
	slice, err := h.repo.GetQueryMetricsSlice(ctx, start, end, serverID, sampleID)
	if err != nil {
		return nil, err
	}
	if len(slice) == 0 {
		return []*common_domain.QueryMetric{}, nil
	}
	const maxBuckets = 10000
	if interval <= 0 {
		interval = time.Minute
	}
	spanSec := end.Sub(start).Seconds()
	if spanSec < 0 {
		spanSec = 0
	}
	if spanSec > 0 && spanSec/interval.Seconds() >= float64(maxBuckets) {
		interval = time.Duration(spanSec/float64(maxBuckets-1)*float64(time.Second)) + time.Nanosecond
	}
	numBuckets := int(spanSec/interval.Seconds()) + 1
	if numBuckets < 1 {
		numBuckets = 1
	}
	if numBuckets > maxBuckets {
		numBuckets = maxBuckets
	}

	buckets := make([]*common_domain.QueryMetric, numBuckets)
	for _, m := range slice {
		if m == nil {
			continue
		}
		idx := int(m.CollectionTime.Sub(start).Seconds() / interval.Seconds())
		if idx < 0 {
			idx = 0
		}
		if idx >= numBuckets {
			idx = numBuckets - 1
		}
		mm := buckets[idx]
		if mm == nil {
			countersCopy := make(map[string]int64, len(m.Counters))
			for k, v := range m.Counters {
				countersCopy[k] = v
			}
			buckets[idx] = &common_domain.QueryMetric{
				QueryHash:         m.QueryHash,
				Database:          m.Database,
				LastExecutionTime: m.LastExecutionTime,
				LastElapsedTime:   m.LastElapsedTime,
				Counters:          countersCopy,
				CollectionTime:    start.Add(time.Duration(idx) * interval),
			}
		} else {
			for k, v := range m.Counters {
				mm.Counters[k] += v
			}
			if mm.LastExecutionTime.Before(m.LastExecutionTime) {
				mm.LastExecutionTime = m.LastExecutionTime
				mm.LastElapsedTime = m.LastElapsedTime
			}
		}
	}

	ret := make([]*common_domain.QueryMetric, 0, len(slice))
	for _, b := range buckets {
		if b != nil {
			ret = append(ret, b)
		}
	}
	return ret, nil
}
