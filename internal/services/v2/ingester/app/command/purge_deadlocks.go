package command

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type PurgeDeadlocks struct {
	Start     time.Time
	End       time.Time
	BatchSize int
}

type PurgeDeadlocksHandler struct {
	repo domain.DeadlockRepository
}

func NewPurgeDeadlocksHandler(repo domain.DeadlockRepository) *PurgeDeadlocksHandler {
	return &PurgeDeadlocksHandler{repo: repo}
}

func (h *PurgeDeadlocksHandler) Handle(ctx context.Context, cmd PurgeDeadlocks) error {
	if h.repo == nil {
		return nil
	}
	return h.repo.PurgeDeadlocks(ctx, cmd.Start, cmd.End, cmd.BatchSize)
}
