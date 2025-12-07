package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
)

type Application struct {
	Queries     Queries
	Commands    Commands
	EventRouter *events.EventRouter
}

type Queries struct {
	ReadMetrics     query.ReadMetricsHandler
	ReadSnapshot    query.ReadSnapshotHandler
	GetQueryPlans   query.GetQueryPlansHandler
	GetKnownHandles query.GetKnownPlanHandlesHandler
}

type Commands struct {
	UploadMetrics   command.UploadMetricsHandler
	UploadSnapshot  command.UploadSnapshotHandler
	UploadExecPlans command.UploadExecPlansHandler
}

func NewApplication(samplesReader domain.SamplesReader, reader domain.QueryMetricsReader,
	client domain.IngestionClient, router *events.EventRouter) *Application {
	return &Application{
		Queries: Queries{
			ReadMetrics:     *query.NewReadMetricsHandler(reader),
			ReadSnapshot:    *query.NewReadSnapshotHandler(samplesReader),
			GetQueryPlans:   *query.NewGetQueryPlansHandler(samplesReader),
			GetKnownHandles: *query.NewGetKnownPlanHandlesHandler(client),
		},
		Commands: Commands{
			UploadMetrics:   *command.NewUploadMetricsHandler(client),
			UploadSnapshot:  *command.NewUploadSnapshotHandler(client),
			UploadExecPlans: *command.NewUploadExecPlansHandler(client),
		},
		EventRouter: router,
	}
}
