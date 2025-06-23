package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"time"
)

type SnapshotSummariesQuery struct {
	Start      time.Time
	End        time.Time
	DatabaseID string
	ServerID   string
}
type ListSnapshotSummariesHandler struct {
	repo domain.SampleRepository
}

func NewListSnapshotSummariesHandler(repo domain.SampleRepository) ListSnapshotSummariesHandler {
	return ListSnapshotSummariesHandler{repo: repo}
}

func (h *ListSnapshotSummariesHandler) Handle(ctx context.Context, query SnapshotSummariesQuery) ([]common_domain.SnapshotSummary, error) {
	return h.repo.ListSnapshotSummaries(ctx, query.ServerID, query.Start, query.End)
}
