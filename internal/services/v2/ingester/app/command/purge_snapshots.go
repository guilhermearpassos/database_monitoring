package command

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type PurgeSnapshots struct {
	Start     time.Time
	End       time.Time
	BatchSize int
}

type PurgeSnapshotsHandler struct {
	repo domain.SnapshotRepository
}

func NewPurgeSnapshotsHandler(repo domain.SnapshotRepository) *PurgeSnapshotsHandler {
	return &PurgeSnapshotsHandler{repo: repo}
}

func (h *PurgeSnapshotsHandler) Handle(ctx context.Context, cmd PurgeSnapshots) error {
	if cmd.BatchSize < 0 {
		return h.repo.PurgeAllSnapshots(ctx)
	}
	return h.repo.PurgeSnapshots(ctx, cmd.Start, cmd.End, cmd.BatchSize)
}
