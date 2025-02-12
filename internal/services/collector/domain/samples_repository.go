package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type SampleRepository interface {
	StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error
	StoreExecutionPlans(ctx context.Context, snapshot []*common_domain.ExecutionPlan) error
	GetKnownPlanHandles(ctx context.Context, server *common_domain.ServerMeta) ([]string, error)

	//ListServers(ctx context.Context, start time.Time, end time.Time) ([]common_domain.ServerSummary, error)
	//ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int) ([]DataBaseSnapshot, int, error)
}

type QueryMetricsRepository interface {
	StoreQueryMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, timestamp time.Time) error
}
