package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type StoreSnapShotSamplesHandler struct {
	repo domain.SampleRepository
}

func NewStoreSnapShotSamplesHandler(repo domain.SampleRepository) StoreSnapShotSamplesHandler {
	return StoreSnapShotSamplesHandler{repo: repo}
}

func (h *StoreSnapShotSamplesHandler) Handle(ctx context.Context, snapID string, samples []*common_domain.QuerySample) error {
	return h.repo.StoreSnapshotSamples(ctx, snapID, samples)

}
