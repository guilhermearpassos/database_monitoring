package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type StoreExecutionPlansHandler struct {
	repo domain.SampleRepository
}

func NewStoreExecutionPlansHandler(repo domain.SampleRepository) StoreExecutionPlansHandler {
	return StoreExecutionPlansHandler{repo: repo}
}

func (h *StoreExecutionPlansHandler) Handle(ctx context.Context, plans []*common_domain.ExecutionPlan) error {
	return h.repo.StoreExecutionPlans(ctx, plans)

}
