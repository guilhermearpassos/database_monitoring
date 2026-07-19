package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type GetQueryPlansHandler struct {
	reader domain.SamplesReader
}

func NewGetQueryPlansHandler(reader domain.SamplesReader) *GetQueryPlansHandler {
	return &GetQueryPlansHandler{reader: reader}
}

func (h GetQueryPlansHandler) Handle(ctx context.Context, handles []string, server common_domain.ServerMeta) (map[string]*common_domain.ExecutionPlan, error) {
	return h.reader.GetPlanHandles(ctx, handles, server)
}
