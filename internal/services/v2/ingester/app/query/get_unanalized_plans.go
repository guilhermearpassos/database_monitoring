package query

import (
	"context"
	"fmt"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type GetUnanalizedPlansHandler struct {
	repo domain.SnapshotRepository
}

func NewGetUnanalizedPlansHandler(repo domain.SnapshotRepository) *GetUnanalizedPlansHandler {
	return &GetUnanalizedPlansHandler{repo: repo}
}

// GetUnanalized returns domain status and missing chunk sequences.
func (r GetUnanalizedPlansHandler) Handle(ctx context.Context, batchSize int) ([]*common_domain.ExecutionPlan, error) {
	handles, err := r.repo.GetUnanalizedPlans(ctx, batchSize)
	if err != nil {
		return nil, fmt.Errorf("get missing plans error: %w", err)
	}
	return handles, nil
}
