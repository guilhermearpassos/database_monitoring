package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/domain"
	"time"
)

type SnapshotsQuery struct {
	Start      time.Time
	End        time.Time
	PageNumber int
	PageSize   int
	DatabaseID string
}

type ListSnapshotsHandler struct {
	repo domain.SampleRepository
}

func NewListSnapshotsHandler(repo domain.SampleRepository) ListSnapshotsHandler {
	return ListSnapshotsHandler{repo: repo}
}

func (h ListSnapshotsHandler) Handle(ctx context.Context, query SnapshotsQuery) ([]common_domain.DataBaseSnapshot, int, error) {
	return h.repo.ListSnapshots(ctx, query.DatabaseID, query.Start, query.End, query.PageNumber, query.PageSize)
}
