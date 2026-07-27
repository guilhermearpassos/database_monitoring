package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/adapters/repository"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/adapters/state"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app/query"
)

type Command struct {
	SaveSnapshot *command.SaveSnapshotStreamHandler
}
type Query struct {
	GetMissingChunks *query.GetMissingChunksHandler
}

// Application wraps command and query handlers for the ingester use-cases.
// Transport adapters (gRPC, HTTP, inproc) should depend on this type, not on
// individual command/query structs.
type Application struct {
	Command Command
	Query   Query
}

// NewApplication builds an application with in-memory store and noop repository (default dev/testing).
func NewApplication() Application {
	store := state.NewMemoryStore()
	repo := repository.NewNoopRepo()
	return NewApplicationWithAdapters(store, repo)
}

// NewApplicationWithAdapters allows wiring custom store and repository implementations.
func NewApplicationWithAdapters(store domain.SessionStore, repo domain.SnapshotRepository) Application {
	return Application{
		Command: Command{
			SaveSnapshot: command.NewSaveSnapshotStreamHandler(store, repo),
		},
		Query: Query{
			GetMissingChunks: query.NewGetMissingChunksHandler(store),
		},
	}
}
