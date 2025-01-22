package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type StoreSnapShotHandler struct {
	repo domain.SampleRepository
}

func NewStoreSnapShotHandler(repo domain.SampleRepository) StoreSnapShotHandler {
	return StoreSnapShotHandler{repo: repo}
}

func (h *StoreSnapShotHandler) Handle(ctx context.Context, snapshot common_domain.DataBaseSnapshot) error {
	return h.repo.StoreSnapshot(ctx, snapshot)

}
