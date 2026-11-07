package query

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type GetMissingChunksHandler struct {
	store domain.SessionStore
}

func NewGetMissingChunksHandler(store domain.SessionStore) *GetMissingChunksHandler {
	return &GetMissingChunksHandler{store: store}
}

// GetMissing returns domain status and missing chunk sequences.
func (r GetMissingChunksHandler) Handle(ctx context.Context, snapshotID string) (domain.UploadStatus, []uint32, error) {
	if snapshotID == "" {
		return domain.UploadStatusUnknown, nil, nil
	}
	st, miss, err := r.store.GetMissing(snapshotID)
	if err != nil {
		return domain.UploadStatusUnknown, nil, err
	}
	return st, miss, nil
}
