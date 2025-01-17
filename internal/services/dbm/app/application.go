package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/domain"
)

type Application struct {
	Queries Queries
}

type Queries struct {
	ListDatabases query.ListDatabasesHandler
	ListSnapshots query.ListSnapshotsHandler
}

func NewApplication(repo domain.SampleRepository) Application {
	return Application{
		Queries: Queries{
			ListDatabases: query.NewListDatabasesHandler(repo),
			ListSnapshots: query.NewListSnapshotsHandler(repo),
		},
	}
}
