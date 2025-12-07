package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type UploadMetricsHandler struct {
	client domain.IngestionClient
	tracer trace.Tracer
}

func NewUploadMetricsHandler(client domain.IngestionClient) *UploadMetricsHandler {
	return &UploadMetricsHandler{client: client, tracer: otel.Tracer("UploadMetrics")}
}

func (h UploadMetricsHandler) Handle(ctx context.Context, metrics []*common_domain.QueryMetric, server common_domain.ServerMeta, sampleTime time.Time) error {
	return h.client.IngestMetrics(ctx, metrics, server, sampleTime)
}
