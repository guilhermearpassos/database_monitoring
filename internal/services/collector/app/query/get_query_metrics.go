package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type GetQueryMetricsHandler struct {
	repo domain.QueryMetricsRepository
}

func NewGetQueryMetricsHandler(repo domain.QueryMetricsRepository) GetQueryMetricsHandler {
	return GetQueryMetricsHandler{repo: repo}
}

func (h GetQueryMetricsHandler) Handle(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string) (*common_domain.QueryMetric, error) {
	return h.repo.GetQueryMetrics(ctx, start, end, serverID, sampleID)
}
