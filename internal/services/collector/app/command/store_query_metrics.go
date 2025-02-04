package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type StoreQueryMetricsHandler struct {
	repo domain.QueryMetricsRepository
}

func NewStoreQueryMetricsHandler(repo domain.QueryMetricsRepository) StoreQueryMetricsHandler {
	return StoreQueryMetricsHandler{repo: repo}
}

func (h *StoreQueryMetricsHandler) Handle(ctx context.Context, metrics []*common_domain.QueryMetric, timestamp time.Time) error {
	return h.repo.StoreQueryMetrics(ctx, metrics, timestamp)

}
