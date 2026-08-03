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

func (h *SetPlanAnalisysHandler) Handle(ctx context.Context, planHandle string, planAnalisysResults domain.PlanAnalisysResults) error {
	return h.repo.SetPlanAnalisys(ctx, planHandle, planAnalisysResults)

}
