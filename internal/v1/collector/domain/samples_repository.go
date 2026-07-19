package domain

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type SampleRepository interface {
	StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error
	StoreSnapshotSamples(ctx context.Context, snapID string, samples []*common_domain.QuerySample) error
	StoreExecutionPlans(ctx context.Context, snapshot []*common_domain.ExecutionPlan) error
	GetKnownPlanHandles(ctx context.Context, server *common_domain.ServerMeta, pageNumber int, pageSize int) ([]string, int, error)
	ListServers(ctx context.Context, start time.Time, end time.Time) ([]ServerSummary, error)
	ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error)
	GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error)
	GetExecutionPlan(ctx context.Context, planHandle string, server *common_domain.ServerMeta) (*common_domain.ExecutionPlan, error)
	GetQuerySample(ctx context.Context, snapID string, sampleID string) (*common_domain.QuerySample, error)
	ListSnapshotSummaries(ctx context.Context, serverID string, start time.Time, end time.Time) ([]common_domain.SnapshotSummary, error)
	PurgeSnapshots(ctx context.Context, start time.Time, end time.Time, size int) error
	PurgeQueryPlans(ctx context.Context, batchSize int) error
	PurgeAllQueryPlans(ctx context.Context) error
	PurgeAllSnapshots(ctx context.Context) error
}

type QueryMetricsRepository interface {
	StoreQueryMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, serverMeta common_domain.ServerMeta, timestamp time.Time) error
	ListQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string) ([]*common_domain.QueryMetric, error)
	GetQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string) (*common_domain.QueryMetric, error)
	GetQueryMetricsSlice(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string) ([]*common_domain.QueryMetric, error)
	PurgeQueryMetrics(ctx context.Context, start time.Time, end time.Time, batchSize int) error
	PurgeAllQueryMetrics(ctx context.Context) error
}

type WarningsRepository interface {
	StoreWarnings(ctx context.Context, warnings []*common_domain.Warning, serverMeta common_domain.ServerMeta) error
	GetKnownWarnings(ctx context.Context, serverID string, pageSize int, pageNumber int) ([]*common_domain.Warning, error)
	GetWarningsByType(ctx context.Context, serverID string, warningType string) ([]*common_domain.Warning, error)
	DeleteWarning(ctx context.Context, serverID string, warningName string) error
	DeleteWarningsByType(ctx context.Context, serverID string, warningType string) error
	GetWarningStats(ctx context.Context) (map[string]map[string]int, error)
	BatchStoreWarnings(ctx context.Context, warnings []*common_domain.Warning, serverMeta common_domain.ServerMeta, batchSize int) error
}
