package domain

import dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"

// SnapshotRepository persists decoded samples for a snapshot upload.
// Implementations should be idempotent at least per (snapshotID, seq) to avoid
// duplicates when chunks are retried.
type SnapshotRepository interface {
	SaveSamples(snapshotID string, seq uint32, samples []*dbmv1.QuerySample) error
	FinalizeSnapshot(snapshotID string) error
}
