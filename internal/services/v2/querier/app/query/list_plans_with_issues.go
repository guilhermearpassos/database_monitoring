package query

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
)

type ListPlansWithIssuesQuery struct {
	Start      time.Time
	End        time.Time
	ServerID   string
	PageNumber int
	PageSize   int
}
type ListPlansWithIssuesHandler struct {
	repo domain.SampleRepository
}

func NewListPlansWithIssuesHandler(repo domain.SampleRepository) ListPlansWithIssuesHandler {
	return ListPlansWithIssuesHandler{repo: repo}
}

func (h *ListPlansWithIssuesHandler) Handle(ctx context.Context, query ListPlansWithIssuesQuery) ([]*domain.PlanWithIssue, error) {
	return h.repo.ListPlansWithIssues(ctx, query.ServerID, query.Start, query.End, query.PageNumber, query.PageSize)
}
