package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type SampleRepository interface {
	ListServers(ctx context.Context, start time.Time, end time.Time) ([]ServerSummary, error)
	ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error)
	GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error)
	GetExecutionPlan(ctx context.Context, planHandle []byte, server *common_domain.ServerMeta) (*common_domain.ExecutionPlan, error)
}

type QueryMetricsRepository interface {
	GetQueryMetrics(ctx context.Context, start time.Time, end time.Time) ([]*common_domain.QueryMetric, error)
}
