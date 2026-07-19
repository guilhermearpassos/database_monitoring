package query

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
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
	points := end.Sub(start).Seconds()/interval.Seconds() + 1
	ret := make([]*common_domain.QueryMetric, int(points))
	for _, m := range slice {
		m.Text = ""
		idx := int(m.CollectionTime.Sub(start).Seconds() / interval.Seconds())
		mm := ret[idx]
		if mm == nil {
			ret[idx] = m
		} else {
			for k := range m.Counters {
				mm.Counters[k] += mm.Counters[k]
			}
			if mm.LastExecutionTime.Before(m.LastExecutionTime) {
				mm.LastExecutionTime = m.LastExecutionTime
				mm.LastElapsedTime = m.LastElapsedTime
			}
		}

	}
	return ret, nil
}
