package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ReadSnapshotHandler struct {
	reader domain.SamplesReader
	tracer trace.Tracer
}

func NewReadSnapshotHandler(reader domain.SamplesReader) *ReadSnapshotHandler {
	return &ReadSnapshotHandler{reader: reader, tracer: otel.Tracer("ReadSnapshot")}
}

func (h ReadSnapshotHandler) Handle(ctx context.Context, serverData common_domain.ServerMeta) ([]*common_domain.DataBaseSnapshot, error) {
	return h.reader.TakeSnapshot(ctx, serverData)
}
