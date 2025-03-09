package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type SnapshotsQuery struct {
	Start      time.Time
	End        time.Time
	PageNumber int
	PageSize   int
	DatabaseID string
	ServerID   string
}

type ListSnapshotsHandler struct {
	repo   domain.SampleRepository
	tracer trace.Tracer
}

func NewListSnapshotsHandler(repo domain.SampleRepository) ListSnapshotsHandler {
	return ListSnapshotsHandler{repo: repo, tracer: otel.Tracer("ListSnapshotsHandler")}
}

func (h ListSnapshotsHandler) Handle(ctx context.Context, query SnapshotsQuery) ([]common_domain.DataBaseSnapshot, int, error) {
	ctx, span := h.tracer.Start(ctx, "ListSnapshots")
	defer span.End()
	return h.repo.ListSnapshots(ctx, query.DatabaseID, query.Start, query.End, query.PageNumber, query.PageSize, query.ServerID)
}
