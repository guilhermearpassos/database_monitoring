package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type SampleRepository interface {
	StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error
	//ListServers(ctx context.Context, start time.Time, end time.Time) ([]common_domain.ServerSummary, error)
	//ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int) ([]DataBaseSnapshot, int, error)
}
