package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type SamplesReader interface {
	TakeSnapshot(ctx context.Context, server common_domain.ServerMeta) ([]*common_domain.DataBaseSnapshot, error)
	GetPlanHandles(ctx context.Context, handles []string, server common_domain.ServerMeta) (map[string]*common_domain.ExecutionPlan, error)
}
type QueryMetricsReader interface {
	CollectMetrics(ctx context.Context, server common_domain.ServerMeta) ([]*common_domain.QueryMetric, error)
}
