package background_agent

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type MetricsCollector struct {
	app    app.Application
	tracer trace.Tracer
}

func NewMetricsCollector(app app.Application) *MetricsCollector {
	return &MetricsCollector{app: app, tracer: otel.Tracer("MetricsCollector")}
}

func (m MetricsCollector) TakeSnapshot(ctx context.Context, server common_domain.ServerMeta, databases []string) (err error) {
	ctx, span := m.tracer.Start(ctx, "MetricsSnapshot")
	span.SetAttributes(
		attribute.String("dbm.server.host", server.Host),
		attribute.String("dbm.server.type", server.Type),
		attribute.Int("dbm.databases.count", len(databases)),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	telemetry.Info(ctx, "starting metrics read", "server", server.Host, "type", server.Type, "dbs", databases)
	sampleTime := time.Now()
	metrics, err := m.app.Queries.ReadMetrics.Handle(ctx, server, databases)
	if err != nil {
		telemetry.Error(ctx, err, "read metrics failed", "server", server.Host)
		return err
	}
	span.SetAttributes(attribute.Int("dbm.metrics.count", len(metrics)))
	err = m.app.Commands.UploadMetrics.Handle(ctx, metrics, server, sampleTime)
	if err != nil {
		telemetry.Error(ctx, err, "upload metrics failed", "server", server.Host, "metrics", len(metrics))
		return err
	}
	m.app.EventRouter.Route(events.MetricsSnapshotTaken{Metrics: metrics, Ctx: ctx})
	telemetry.Info(ctx, "metrics snapshot completed", "server", server.Host, "metrics", len(metrics))
	return nil
}

func (m MetricsCollector) Run(ctx context.Context, server common_domain.ServerMeta, databases []string, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	defer func() {
		if rec := recover(); rec != nil {
			telemetry.Error(ctx, fmt.Errorf("%v", rec), "metrics collector panic recovered", "server", server.Host)
		}
	}()
	for {
		err := m.TakeSnapshot(ctx, server, databases)
		if err != nil {
			telemetry.Error(ctx, err, "metrics collector iteration failed", "server", server.Host)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// next iteration
		}
	}
}
