package repository

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
)

type NoopSampleRepo struct {
}

func NewNoopSampleRepo() *NoopSampleRepo {
	return &NoopSampleRepo{}
}

var _ domain.SampleRepository = (*NoopSampleRepo)(nil)

func (n NoopSampleRepo) ListServers(ctx context.Context, start time.Time, end time.Time) ([]domain.ServerSummary, error) {
	//TODO implement me
	panic("implement me")
}

func (n NoopSampleRepo) ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error) {
	//TODO implement me
	panic("implement me")
}

func (n NoopSampleRepo) GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (n NoopSampleRepo) ListSnapshotSummaries(ctx context.Context, serverID string, start time.Time, end time.Time) ([]common_domain.SnapshotSummary, error) {
	//TODO implement me
	panic("implement me")
}
