package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"time"
)

type PurgeQueryMetrics struct {
	Start     time.Time
	End       time.Time
	BatchSize int
}

type PurgeQueryMetricsHandler struct {
	repo domain.QueryMetricsRepository
}

func NewPurgeQueryMetricsHandler(repo domain.QueryMetricsRepository) PurgeQueryMetricsHandler {
	return PurgeQueryMetricsHandler{repo: repo}
}

func (h *PurgeQueryMetricsHandler) Handle(ctx context.Context, cmd PurgeQueryMetrics) error {
	if cmd.BatchSize <= 0 {
		return h.repo.PurgeAllQueryMetrics(ctx)
	}
	return h.repo.PurgeQueryMetrics(ctx, cmd.Start, cmd.End, cmd.BatchSize)
}
