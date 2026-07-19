package query

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type GetKnownWarningsHandler struct {
	repo domain.WarningsRepository
}

func NewGetKnownWarningsHandler(repo domain.WarningsRepository) GetKnownWarningsHandler {
	return GetKnownWarningsHandler{repo: repo}
}

func (h GetKnownWarningsHandler) Handle(ctx context.Context, serverID string, pageSize int, pageNumber int) ([]*common_domain.Warning, error) {
	return h.repo.GetKnownWarnings(ctx, serverID, pageSize, pageNumber)
}
