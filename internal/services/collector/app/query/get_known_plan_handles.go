package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type GetKnownPlanHandlesHandler struct {
	repo domain.SampleRepository
}

func NewGetKnownPlanHandlesHandler(repo domain.SampleRepository) GetKnownPlanHandlesHandler {
	return GetKnownPlanHandlesHandler{repo: repo}
}

func (h *GetKnownPlanHandlesHandler) Handle(ctx context.Context, snapshot *common_domain.ServerMeta) ([][]byte, error) {
	return h.repo.GetKnownPlanHandles(ctx, snapshot)

}
