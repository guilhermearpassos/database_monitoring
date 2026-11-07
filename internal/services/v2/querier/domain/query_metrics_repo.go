package domain

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type QueryMetricsRepository interface {
	ListQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string) ([]*common_domain.QueryMetric, error)
	GetQueryMetricsSlice(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string) ([]*common_domain.QueryMetric, error)
}
