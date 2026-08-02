package command

import (
	"context"
	"fmt"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type SaveExecutionPlansStreamHandler struct {
	repo domain.SnapshotRepository
}

func NewSaveExecutionPlansStreamHandler(repo domain.SnapshotRepository) *SaveExecutionPlansStreamHandler {
	return &SaveExecutionPlansStreamHandler{repo: repo}
}

func (s SaveExecutionPlansStreamHandler) Handle(ctx context.Context, it domain.UploadIterator[domain.PlanUploadMessage]) error {
	var (
		headerSeen   bool
		finalizeSeen bool
	)

	for {
		msg, err := it.Next(ctx)
		if err != nil {
			return fmt.Errorf("stream read: %w", err)
		}
		if msg == nil {
			break // EOF
		}

		if msg.Header != nil {
			if headerSeen {
				return domain.ErrDuplicateHeader
			}
			headerSeen = true
			continue
		}

		if msg.Chunk != nil {
			if !headerSeen {
				return domain.ErrChunkBeforeHeader
			}
			ch := msg.Chunk
			if ch.ChunkSeq == 0 {
				return domain.ErrInvalidChunkSeq
			}
			if err := s.repo.SaveExecutionPlans(ctx, ch.ExecutionPlans); err != nil {
				return fmt.Errorf("repo save chunk %d: %w", ch.ChunkSeq, err)
			}
			continue
		}

		if msg.Finalize != nil {
			if !headerSeen {
				return domain.ErrMissingHeader
			}
			if finalizeSeen {
				return domain.ErrDuplicateFinalize
			}
			finalizeSeen = true
			continue
		}

		return fmt.Errorf("empty upload message payload")
	}

	return nil
}
