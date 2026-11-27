package domain

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type DeadlockRepository interface {
	ListDeadlocks(ctx context.Context, start time.Time, end time.Time, host string, databases []string) ([]*common_domain.DeadlockReport, error)
}
