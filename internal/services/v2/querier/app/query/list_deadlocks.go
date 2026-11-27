package query

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
)

type ListDeadlocksHandler struct {
	repo domain.DeadlockRepository
}

func NewListDeadlocksHandler(repo domain.DeadlockRepository) ListDeadlocksHandler {
	return ListDeadlocksHandler{repo: repo}
}

func (h ListDeadlocksHandler) Handle(
	ctx context.Context,
	start time.Time,
	end time.Time,
	host string,
	databases []string,
) ([]*common_domain.DeadlockReport, error) {
	if h.repo == nil {
		return []*common_domain.DeadlockReport{}, nil
	}
	return h.repo.ListDeadlocks(ctx, start, end, host, databases)
}
