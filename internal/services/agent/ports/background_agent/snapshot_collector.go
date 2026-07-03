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

type SnapshotCollector struct {
	app    app.Application
	tracer trace.Tracer
}

func NewSnapshotCollector(app app.Application) *SnapshotCollector {
	return &SnapshotCollector{app: app, tracer: otel.Tracer("SnapshotCollector")}
}

func (m SnapshotCollector) TakeSnapshot(ctx context.Context, server common_domain.ServerMeta, databases []string) (err error) {
	ctx, span := m.tracer.Start(ctx, "SamplesSnapshot")
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
	telemetry.Info(ctx, "starting snapshot read", "server", server.Host, "type", server.Type, "dbs", databases)
	snapshots, err := m.app.Queries.ReadSnapshot.Handle(ctx, server, databases)
	if err != nil {
		telemetry.Error(ctx, err, "read snapshot failed", "server", server.Host, "type", server.Type)
		return fmt.Errorf("reading metrics: %w", err)
	}
	span.SetAttributes(attribute.Int("dbm.snapshots.count", len(snapshots)))
	for _, snap := range snapshots {
		if snap == nil || snap.SnapInfo.ID == "" {
			continue
		}
		span.AddEvent("upload_snapshot", trace.WithAttributes(
			attribute.String("dbm.snapshot.id", snap.SnapInfo.ID),
			attribute.Int("dbm.samples.count", len(snap.Samples)),
		))
		err = m.app.Commands.UploadSnapshot.Handle(ctx, snap)
		if err != nil {
			telemetry.Error(ctx, err, "upload snapshot failed", "snapshot_id", snap.SnapInfo.ID, "server", server.Host)
			return fmt.Errorf("uploading metrics: %w", err)
		}
		m.app.EventRouter.Route(events.SampleSnapshotTaken{Snap: snap, Ctx: ctx})
	}
	telemetry.Info(ctx, "snapshot cycle completed", "server", server.Host, "snapshots", len(snapshots))
	return nil
}
func (s SnapshotCollector) Run(ctx context.Context, server common_domain.ServerMeta, databases []string, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	defer func() {
		if rec := recover(); rec != nil {
			telemetry.Error(ctx, fmt.Errorf("%v", rec), "snapshot collector panic recovered", "server", server.Host)
		}
	}()
	for {
		err := s.TakeSnapshot(ctx, server, databases)
		if err != nil {
			telemetry.Error(ctx, err, "snapshot collector iteration failed", "server", server.Host)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// next iteration
		}
	}
}
