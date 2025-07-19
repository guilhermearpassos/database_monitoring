package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type ListQueryMetricsHandler struct {
	repo domain.QueryMetricsRepository
}

func NewListQueryMetricsHandler(repo domain.QueryMetricsRepository) ListQueryMetricsHandler {
	return ListQueryMetricsHandler{repo: repo}
}

func (h ListQueryMetricsHandler) Handle(ctx context.Context, start time.Time, end time.Time, serverID string) ([]*common_domain.QueryMetric, error) {
	return h.repo.ListQueryMetrics(ctx, start, end, serverID)
}
