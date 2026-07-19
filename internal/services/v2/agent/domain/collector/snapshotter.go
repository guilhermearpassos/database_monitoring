package collector

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type Snapshotter interface {
	TakeSnapshot(ctx context.Context, databases []string) (*common_domain.DataBaseSnapshot, error)
	FetchExecutionPlans(ctx context.Context, handles []string) map[string]*common_domain.ExecutionPlan
}
