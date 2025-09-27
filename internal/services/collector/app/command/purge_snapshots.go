package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"time"
)

type PurgeSnapshots struct {
	Start     time.Time
	End       time.Time
	BatchSize int
}

type PurgeSnapshotsHandler struct {
	repo domain.SampleRepository
}

func NewPurgeSnapshotsHandler(repo domain.SampleRepository) PurgeSnapshotsHandler {
	return PurgeSnapshotsHandler{repo: repo}
}

func (h *PurgeSnapshotsHandler) Handle(ctx context.Context, cmd PurgeSnapshots) error {
	return h.repo.PurgeSnapshots(ctx, cmd.Start, cmd.End, cmd.BatchSize)
}
