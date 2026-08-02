package query

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type GetMissingPlansHandler struct {
	repo domain.SnapshotRepository
}

func NewGetMissingPlansHandler(repo domain.SnapshotRepository) *GetMissingPlansHandler {
	return &GetMissingPlansHandler{repo: repo}
}

// GetMissing returns domain status and missing chunk sequences.
func (r GetMissingPlansHandler) Handle(ctx context.Context, server common_domain.ServerMeta, start time.Time, end time.Time) ([]string, error) {
	handles, err := r.repo.GetMissingPlans(ctx, server, start, end)
	if err != nil {
		return nil, fmt.Errorf("get missing plans error: %w", err)
	}
	return handles, nil
}
