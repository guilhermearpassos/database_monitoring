package command

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type StoreQueryMetricsHandler struct {
	repo domain.QueryMetricsRepository
}

func NewStoreQueryMetricsHandler(repo domain.QueryMetricsRepository) *StoreQueryMetricsHandler {
	return &StoreQueryMetricsHandler{repo: repo}
}

func (h *StoreQueryMetricsHandler) Handle(ctx context.Context, metrics []*common_domain.QueryMetric, serverMeta common_domain.ServerMeta, timestamp time.Time) error {
	return h.repo.StoreQueryMetrics(ctx, metrics, serverMeta, timestamp)

}
