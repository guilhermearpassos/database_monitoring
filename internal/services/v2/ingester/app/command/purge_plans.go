package command

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type PurgePlans struct {
	BatchSize int
}

type PurgePlansHandler struct {
	repo domain.SnapshotRepository
}

func NewPurgePlansHandler(repo domain.SnapshotRepository) *PurgePlansHandler {
	return &PurgePlansHandler{repo: repo}
}

func (h *PurgePlansHandler) Handle(ctx context.Context, cmd PurgePlans) error {
	return h.repo.PurgeUnboundedPlans(ctx, cmd.BatchSize)
}
