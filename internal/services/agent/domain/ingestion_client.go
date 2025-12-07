package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type IngestionClient interface {
	IngestMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, server common_domain.ServerMeta, timestamp time.Time) error
	IngestSnapshot(ctx context.Context, snapshot *common_domain.DataBaseSnapshot) error
	IngestExecPlans(ctx context.Context, executionPlans map[string]*common_domain.ExecutionPlan, server common_domain.ServerMeta) error
	GetKnownPlanHandles(ctx context.Context, server common_domain.ServerMeta) (map[string]struct{}, error)
}
