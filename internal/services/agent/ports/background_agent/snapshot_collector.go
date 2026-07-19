package background_agent

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
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
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		}
		span.End()
	}()
	snapshots, err := m.app.Queries.ReadSnapshot.Handle(ctx, server, databases)
	if err != nil {
		return fmt.Errorf("reading metrics: %w", err)
	}
	for _, snap := range snapshots {

		err = m.app.Commands.UploadSnapshot.Handle(ctx, snap)
		if err != nil {
			return fmt.Errorf("uploading metrics: %w", err)
		}
		m.app.EventRouter.Route(events.SampleSnapshotTaken{Snap: snap, Ctx: ctx})
	}
	return nil
}
func (s SnapshotCollector) Run(ctx context.Context, server common_domain.ServerMeta, databases []string, interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		err := s.TakeSnapshot(ctx, server, databases)
		if err != nil {
			fmt.Printf("taking snapshot %s: %s\n", server.Host, err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			break

		}
	}
}
