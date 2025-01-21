package domain

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type DataBaseReader interface {
	TakeSnapshot(ctx context.Context) ([]*common_domain.DataBaseSnapshot, error)
}
