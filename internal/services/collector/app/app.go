package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Queries struct {
	GetKnownPlanHandlesHandler query.GetKnownPlanHandlesHandler
	GetQuerySampleDetails      query.GetQuerySampleDetailsHandler
	GetSnapshot                query.GetSnapshotHandler
	ListQueryMetrics           query.ListQueryMetricsHandler
	ListServerSummary          query.ListServerSummaryHandler
	ListSnapshots              query.ListSnapshotsHandler
	ListSnapshotSummaries      query.ListSnapshotSummariesHandler
	GetQueryMetrics            query.GetQueryMetricsHandler
	GetKnownWarnings           query.GetKnownWarningsHandler
	GetQueryMetricsSlice       query.GetQueryMetricsSliceHandler
}

type Commands struct {
	StoreSnapshot        command.StoreSnapShotHandler
	StoreQueryMetrics    command.StoreQueryMetricsHandler
	StoreExecutionPlans  command.StoreExecutionPlansHandler
	PurgeQueryMetrics    command.PurgeQueryMetricsHandler
	StoreSnapshotSamples command.StoreSnapShotSamplesHandler
	PurgeSnapshots       command.PurgeSnapshotsHandler
	PurgeQueryPlans      command.PurgeQueryPlansHandler
	StoreWarnings        command.StoreWarningsHandler
}

func NewApplication(repo domain.SampleRepository, queryMetricsRepo domain.QueryMetricsRepository, warnRepo domain.WarningsRepository) *Application {
	return &Application{
		Commands: Commands{StoreSnapshot: command.NewStoreSnapShotHandler(repo),
			StoreQueryMetrics:    command.NewStoreQueryMetricsHandler(queryMetricsRepo),
			StoreExecutionPlans:  command.NewStoreExecutionPlansHandler(repo),
			PurgeQueryMetrics:    command.NewPurgeQueryMetricsHandler(queryMetricsRepo),
			StoreSnapshotSamples: command.NewStoreSnapShotSamplesHandler(repo),
			PurgeSnapshots:       command.NewPurgeSnapshotsHandler(repo),
			PurgeQueryPlans:      command.NewPurgeQueryPlansHandler(repo),
			StoreWarnings:        command.NewStoreWarningsHandler(warnRepo),
		},
		Queries: Queries{
			GetKnownPlanHandlesHandler: query.NewGetKnownPlanHandlesHandler(repo),
			GetQuerySampleDetails:      query.NewGetQuerySampleDetailsHandler(repo),
			GetSnapshot:                query.NewGetSnapshotHandler(repo),
			ListQueryMetrics:           query.NewListQueryMetricsHandler(queryMetricsRepo),
			ListServerSummary:          query.NewListServerSummaryHandler(repo),
			ListSnapshots:              query.NewListSnapshotsHandler(repo),
			ListSnapshotSummaries:      query.NewListSnapshotSummariesHandler(repo),
			GetQueryMetrics:            query.NewGetQueryMetricsHandler(queryMetricsRepo),
			GetKnownWarnings:           query.NewGetKnownWarningsHandler(warnRepo),
			GetQueryMetricsSlice:       query.NewGetQueryMetricsSliceHandler(queryMetricsRepo),
		},
	}
}
