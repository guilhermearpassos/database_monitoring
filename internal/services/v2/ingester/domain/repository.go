package domain

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// SnapshotRepository persists decoded samples for a snapshot upload.
// Implementations should be idempotent at least per (snapshotID, seq) to avoid
// duplicates when chunks are retried.
type SnapshotRepository interface {
	// EnsureSnapshot guarantees a row exists for the given snapshot header in the v1 schema.
	// It should upsert target by (host,type) and insert snapshot by external id (f_id) if missing.
	EnsureSnapshot(ctx context.Context, header SnapshotHeader) error
	// SaveSamples stores a chunk worth of samples for the given snapshot id and sequence.
	SaveSamples(ctx context.Context, snapshotID string, seq uint32, samples []*dbmv1.QuerySample) error
	// FinalizeSnapshot performs any finalize bookkeeping. Can be a no-op under v1 schema.
	FinalizeSnapshot(ctx context.Context, snapshotID string) error

	SaveExecutionPlans(ctx context.Context, executionPlan []*common_domain.ExecutionPlan) error
	GetMissingPlans(ctx context.Context, server common_domain.ServerMeta, start time.Time, end time.Time) ([]string, error)
}
