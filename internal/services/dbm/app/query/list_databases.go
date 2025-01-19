package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/domain"
	"time"
)

type ListDatabasesHandler struct {
	repo domain.SampleRepository
}

func NewListDatabasesHandler(repo domain.SampleRepository) ListDatabasesHandler {
	return ListDatabasesHandler{repo: repo}
}

func (h ListDatabasesHandler) Handle(ctx context.Context, start time.Time, end time.Time) ([]domain.ServerMeta, error) {
	return h.repo.ListServers(ctx, start, end)
}
