package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type GetKnownPlanHandlesHandler struct {
	client domain.IngestionClient
}

func NewGetKnownPlanHandlesHandler(client domain.IngestionClient) *GetKnownPlanHandlesHandler {
	return &GetKnownPlanHandlesHandler{client: client}
}

func (h GetKnownPlanHandlesHandler) Handle(ctx context.Context, server common_domain.ServerMeta) (map[string]struct{}, error) {
	return h.client.GetKnownPlanHandles(ctx, server)
}
