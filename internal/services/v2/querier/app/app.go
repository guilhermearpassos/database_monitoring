package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Queries struct {
	GetSnapshot           query.GetSnapshotHandler
	ListServerSummary     query.ListServerSummaryHandler
	ListSnapshotSummaries query.ListSnapshotSummariesHandler
	GetQueryMetricsSlice  query.GetQueryMetricsSliceHandler
	ListQueryMetrics      query.ListQueryMetricsHandler
	GetQuerySampleDetails query.GetQuerySampleDetailsHandler
	ListPlansWithIssues   query.ListPlansWithIssuesHandler
	ListDeadlocks         query.ListDeadlocksHandler
}

type Commands struct {
}

func NewApplication(repo domain.SampleRepository, queryRepo domain.QueryMetricsRepository, deadlockRepo ...domain.DeadlockRepository) *Application {
	var dlRepo domain.DeadlockRepository
	if len(deadlockRepo) > 0 {
		dlRepo = deadlockRepo[0]
	} else if r, ok := repo.(domain.DeadlockRepository); ok {
		dlRepo = r
	}
	return &Application{
		Commands: Commands{},
		Queries: Queries{
			GetSnapshot:           query.NewGetSnapshotHandler(repo),
			ListServerSummary:     query.NewListServerSummaryHandler(repo),
			ListSnapshotSummaries: query.NewListSnapshotSummariesHandler(repo),
			GetQueryMetricsSlice:  query.NewGetQueryMetricsSliceHandler(queryRepo),
			ListQueryMetrics:      query.NewListQueryMetricsHandler(queryRepo),
			GetQuerySampleDetails: query.NewGetQuerySampleDetailsHandler(repo),
			ListPlansWithIssues:   query.NewListPlansWithIssuesHandler(repo),
			ListDeadlocks:         query.NewListDeadlocksHandler(dlRepo),
		},
	}
}
