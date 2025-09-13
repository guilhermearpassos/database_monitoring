package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type SamplesReader interface {
	TakeSnapshot(ctx context.Context) ([]*common_domain.DataBaseSnapshot, error)
	GetPlanHandles(ctx context.Context, handles []string, ignoreKnown bool) (map[string]*common_domain.ExecutionPlan, error)
}
type QueryMetricsReader interface {
	CollectMetrics(ctx context.Context) ([]*common_domain.QueryMetric, error)
}
