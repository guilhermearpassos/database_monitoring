package command

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type SetPlanAnalisysHandler struct {
	repo domain.SnapshotRepository
}

func NewSetPlanAnalisysHandler(repo domain.SnapshotRepository) *SetPlanAnalisysHandler {
	return &SetPlanAnalisysHandler{repo: repo}
}

func (h *SetPlanAnalisysHandler) Handle(ctx context.Context, results []domain.PlanAnalisysBatchResult) error {
	return h.repo.SetPlanAnalisysBatch(ctx, results)
}
