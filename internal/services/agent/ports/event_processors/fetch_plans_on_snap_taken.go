package event_processors

import (
	"context"
	"fmt"

	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type PlanFetcher struct {
	app                  app.Application
	in                   chan events.Event
	trace                trace.Tracer
	knownHandlesByServer map[string]map[string]struct{}
}

func NewPlanFetcher(app app.Application) *PlanFetcher {
	return &PlanFetcher{
		app:                  app,
		in:                   make(chan events.Event, 200),
		trace:                otel.Tracer("PlanFetcher"),
		knownHandlesByServer: make(map[string]map[string]struct{}),
	}
}

func (f *PlanFetcher) Run(ctx context.Context) {
	for ev := range f.in {
		snapTakenEvent, ok := ev.(events.SampleSnapshotTaken)
		if !ok {
			continue
		}
		ctx, span := f.trace.Start(ctx, "FetchPlansOnSnapTaken")
		handles := snapTakenEvent.Snap.GetPlanHandles()
		newHandles := make([]string, 0, len(handles))
		if m, found := f.knownHandlesByServer[snapTakenEvent.Snap.SnapInfo.Server.Host]; (!found) || (len(m) == 0) {
			known, err := f.app.Queries.GetKnownHandles.Handle(ctx, snapTakenEvent.Snap.SnapInfo.Server)
			if err != nil {
				fmt.Println(err)
				span.SetStatus(otelcodes.Error, err.Error())
				span.RecordError(err)
				known = make(map[string]struct{})
			}
			f.knownHandlesByServer[snapTakenEvent.Snap.SnapInfo.Server.Host] = known
		}
		for _, handle := range handles {
			if _, found := f.knownHandlesByServer[snapTakenEvent.Snap.SnapInfo.Server.Host][handle]; !found {
				newHandles = append(newHandles, handle)
			}
		}
		if len(newHandles) == 0 {
			span.End()
			continue
		}
		plans, err := f.app.Queries.GetQueryPlans.Handle(ctx, newHandles, snapTakenEvent.Snap.SnapInfo.Server)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
		}
		for k := range plans {
			f.knownHandlesByServer[snapTakenEvent.Snap.SnapInfo.Server.Host][k] = struct{}{}
		}
		err = f.app.Commands.UploadExecPlans.Handle(ctx, plans, snapTakenEvent.Snap.SnapInfo.Server)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
		}
		for _, plan := range plans {
			f.app.EventRouter.Route(events.ExecutionPlanFetched{Plan: plan})
		}
		span.End()
	}
}

func (f *PlanFetcher) Register(router *events.EventRouter) {
	router.Register(events.SampleSnapshotTaken{}.EventName(), f.in, "planFetcher")
}
