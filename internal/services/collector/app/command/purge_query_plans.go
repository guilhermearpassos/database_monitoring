package command

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"time"
)

type PurgeQueryPlans struct {
	Start     time.Time
	End       time.Time
	BatchSize int
}

type PurgeQueryPlansHandler struct {
	repo domain.SampleRepository
}

func NewPurgeQueryPlansHandler(repo domain.SampleRepository) PurgeQueryPlansHandler {
	return PurgeQueryPlansHandler{repo: repo}
}

func (h *PurgeQueryPlansHandler) Handle(ctx context.Context, batchSize int) error {
	return h.repo.PurgeQueryPlans(ctx, batchSize)
}
