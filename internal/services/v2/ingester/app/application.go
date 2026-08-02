package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type Command struct {
	SaveSnapshot       *command.SaveSnapshotStreamHandler
	SaveExecutionPlans *command.SaveExecutionPlansStreamHandler
	StoreQueryMetrics  *command.StoreQueryMetricsHandler
}
type Query struct {
	GetMissingChunks *query.GetMissingChunksHandler
	GetMissingPlans  *query.GetMissingPlansHandler
}

// Application wraps command and query handlers for the ingester use-cases.
// Transport adapters (gRPC, HTTP, inproc) should depend on this type, not on
// individual command/query structs.
type Application struct {
	Command Command
	Query   Query
}

// NewApplicationWithAdapters allows wiring custom store and repository implementations.
func NewApplicationWithAdapters(store domain.SessionStore, repo domain.SnapshotRepository, queryRepo domain.QueryMetricsRepository) Application {
	return Application{
		Command: Command{
			SaveSnapshot:       command.NewSaveSnapshotStreamHandler(store, repo),
			SaveExecutionPlans: command.NewSaveExecutionPlansStreamHandler(repo),
			StoreQueryMetrics:  command.NewStoreQueryMetricsHandler(queryRepo),
		},
		Query: Query{
			GetMissingChunks: query.NewGetMissingChunksHandler(store),
			GetMissingPlans:  query.NewGetMissingPlansHandler(repo),
		},
	}
}
