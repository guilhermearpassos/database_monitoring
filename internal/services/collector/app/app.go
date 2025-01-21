package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
)

type Application struct {
	Commands Commands
}

type Commands struct {
	StoreSnapshot command.StoreSnapShotHandler
}

func NewApplication(repo domain.SampleRepository) *Application {
	return &Application{
		Commands: Commands{StoreSnapshot: command.NewStoreSnapShotHandler(repo)},
	}
}
