package query

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
)

type ListPlansWithIssuesQuery struct {
	Start      time.Time
	End        time.Time
	ServerID   string
	Databases  []string
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
	issues, err := h.repo.ListPlansWithIssues(ctx, query.ServerID, query.Start, query.End, query.PageNumber, query.PageSize)
	if err != nil {
		return nil, fmt.Errorf("listing plans: %w", err)
	}
	ret := issues
	if len(query.Databases) > 0 {
		ret = make([]*domain.PlanWithIssue, 0, len(issues))
		for _, issue := range issues {
			if slices.Contains(query.Databases, issue.Sample.Database.DatabaseName) {
				ret = append(ret, issue)
			}
		}
	}
	return ret, nil
}
