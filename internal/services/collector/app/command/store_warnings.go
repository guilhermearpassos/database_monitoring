package command

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type StoreWarnings struct {
	Warnings   []*common_domain.Warning
	ServerMeta common_domain.ServerMeta
}

type StoreWarningsHandler struct {
	repo domain.WarningsRepository
}

func NewStoreWarningsHandler(repo domain.WarningsRepository) StoreWarningsHandler {
	return StoreWarningsHandler{repo: repo}
}

func (h *StoreWarningsHandler) Handle(ctx context.Context, cmd StoreWarnings) error {
	return h.repo.StoreWarnings(ctx, cmd.Warnings, cmd.ServerMeta)
}
