package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type GetSnapshotHandler struct {
	repo domain.SampleRepository
}

func NewGetSnapshotHandler(repo domain.SampleRepository) GetSnapshotHandler {
	return GetSnapshotHandler{repo: repo}
}

func (h GetSnapshotHandler) Handle(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error) {
	return h.repo.GetSnapshot(ctx, id)
}
