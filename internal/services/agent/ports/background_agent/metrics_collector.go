package background_agent

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type MetricsCollector struct {
	app    app.Application
	tracer trace.Tracer
}

func NewMetricsCollector(app app.Application) *MetricsCollector {
	return &MetricsCollector{app: app, tracer: otel.Tracer("MetricsCollector")}
}

func (m MetricsCollector) TakeSnapshot(ctx context.Context, server common_domain.ServerMeta) (err error) {
	ctx, span := m.tracer.Start(ctx, "MetricsSnapshot")
	defer func() {
		if err != nil {

			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	sampleTime := time.Now()
	metrics, err := m.app.Queries.ReadMetrics.Handle(ctx, server)
	if err != nil {
		return fmt.Errorf("reading metrics: %w", err)
	}
	err = m.app.Commands.UploadMetrics.Handle(ctx, metrics, server, sampleTime)
	if err != nil {
		return fmt.Errorf("uploading metrics: %w", err)
	}
	m.app.EventRouter.Route(events.MetricsSnapshotTaken{Metrics: metrics})
	return nil
}

func (m MetricsCollector) Run(ctx context.Context, server common_domain.ServerMeta, interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		err := m.TakeSnapshot(ctx, server)
		if err != nil {
			fmt.Printf("taking snapshot %s: %s", server.Host, err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			break

		}
	}
}
