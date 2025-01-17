package domain

import (
	"context"
	"time"
)

type SampleRepository interface {
	StoreSnapshot(ctx context.Context, snapshot DataBaseSnapshot) error
	ListDatabases(ctx context.Context, start time.Time, end time.Time) ([]InstrumentedServerMetadata, error)
	ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int) ([]DataBaseSnapshot, int, error)
}
