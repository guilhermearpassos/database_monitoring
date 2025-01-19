package domain

import (
	"context"
	"time"
)

type SampleRepository interface {
	StoreSnapshot(ctx context.Context, snapshot DataBaseSnapshot) error
	ListServers(ctx context.Context, start time.Time, end time.Time) ([]ServerMeta, error)
	ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int) ([]DataBaseSnapshot, int, error)
}
