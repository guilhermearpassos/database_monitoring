package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type UploadExecPlansHandler struct {
	client domain.IngestionClient
	tracer trace.Tracer
}

func NewUploadExecPlansHandler(client domain.IngestionClient) *UploadExecPlansHandler {
	return &UploadExecPlansHandler{client: client, tracer: otel.Tracer("UploadExecPlans")}
}

func (h UploadExecPlansHandler) Handle(ctx context.Context, plan map[string]*common_domain.ExecutionPlan, server common_domain.ServerMeta) (err error) {
	return h.client.IngestExecPlans(ctx, plan, server)
}
