package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type MetricsProcessor interface {
	Run(ctx context.Context)
	Stop()
	QueueSnapshot(snapshot *common_domain.DataBaseSnapshot)
}
