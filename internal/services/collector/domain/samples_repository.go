package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type SampleRepository interface {
	StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error
	StoreExecutionPlans(ctx context.Context, snapshot []*common_domain.ExecutionPlan) error
	GetKnownPlanHandles(ctx context.Context, server *common_domain.ServerMeta) ([][]byte, error)
	ListServers(ctx context.Context, start time.Time, end time.Time) ([]ServerSummary, error)
	ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error)
	GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error)
	GetExecutionPlan(ctx context.Context, planHandle []byte, server *common_domain.ServerMeta) (*common_domain.ExecutionPlan, error)
	GetQuerySample(ctx context.Context, snapID string, sampleID string) (*common_domain.QuerySample, error)
	ListSnapshotSummaries(ctx context.Context, serverID string, start time.Time, end time.Time) ([]common_domain.SnapshotSummary, error)
}

type QueryMetricsRepository interface {
	StoreQueryMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, serverMeta common_domain.ServerMeta, timestamp time.Time) error
	ListQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string) ([]*common_domain.QueryMetric, error)
	GetQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID []byte) (*common_domain.QueryMetric, error)
}
