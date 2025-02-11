package app

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
)

type Application struct {
	Commands Commands
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
	}
}
