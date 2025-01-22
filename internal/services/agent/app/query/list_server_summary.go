package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"time"
)

type ListServerSummaryHandler struct {
	repo domain.SampleRepository
}

func NewListServerSummaryHandler(repo domain.SampleRepository) ListServerSummaryHandler {
	return ListServerSummaryHandler{repo: repo}
}

func (h ListServerSummaryHandler) Handle(ctx context.Context, start time.Time, end time.Time) ([]domain.ServerSummary, error) {
	return h.repo.ListServers(ctx, start, end)
}
