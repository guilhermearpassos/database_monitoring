package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ReadMetricsHandler struct {
	reader domain.QueryMetricsReader
	tracer trace.Tracer
}

func NewReadMetricsHandler(reader domain.QueryMetricsReader) *ReadMetricsHandler {
	return &ReadMetricsHandler{reader: reader, tracer: otel.Tracer("ReadMetrics")}
}

func (h ReadMetricsHandler) Handle(ctx context.Context, serverData common_domain.ServerMeta) ([]*common_domain.QueryMetric, error) {
	return h.reader.CollectMetrics(ctx, serverData)
}
