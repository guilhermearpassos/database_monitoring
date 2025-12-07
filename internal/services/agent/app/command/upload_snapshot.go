package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type UploadSnapshotHandler struct {
	client domain.IngestionClient
	tracer trace.Tracer
}

func NewUploadSnapshotHandler(client domain.IngestionClient) *UploadSnapshotHandler {
	return &UploadSnapshotHandler{client: client, tracer: otel.Tracer("UploadSnapshot")}
}

func (h UploadSnapshotHandler) Handle(ctx context.Context, snapshot *common_domain.DataBaseSnapshot) (err error) {
	return h.client.IngestSnapshot(ctx, snapshot)
}
