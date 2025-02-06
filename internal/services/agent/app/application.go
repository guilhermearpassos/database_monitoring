package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
)

type Application struct {
	Queries Queries
}

type Queries struct {
	ListServerSummary query.ListServerSummaryHandler
	ListSnapshots     query.ListSnapshotsHandler
	GetSnapshot       query.GetSnapshotHandler
	ListQueryMetrics  query.ListQueryMetricsHandler
}

func NewApplication(repo domain.SampleRepository, metricsRepo domain.QueryMetricsRepository) Application {
	return Application{
		Queries: Queries{
			ListServerSummary: query.NewListServerSummaryHandler(repo),
			ListSnapshots:     query.NewListSnapshotsHandler(repo),
			GetSnapshot:       query.NewGetSnapshotHandler(repo),
			ListQueryMetrics:  query.NewListQueryMetricsHandler(metricsRepo),
		},
	}
}
