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
}

type Commands struct {
	StoreSnapshot       command.StoreSnapShotHandler
	StoreQueryMetrics   command.StoreQueryMetricsHandler
	StoreExecutionPlans command.StoreExecutionPlansHandler
}

func NewApplication(repo domain.SampleRepository, queryMetricsRepo domain.QueryMetricsRepository) *Application {
	return &Application{
		Commands: Commands{StoreSnapshot: command.NewStoreSnapShotHandler(repo),
			StoreQueryMetrics:   command.NewStoreQueryMetricsHandler(queryMetricsRepo),
			StoreExecutionPlans: command.NewStoreExecutionPlansHandler(repo),
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
		},
	}
}
