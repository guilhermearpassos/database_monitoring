package command

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

type SaveSnapshotStreamHandler struct {
	store domain.SessionStore
	repo  domain.SnapshotRepository
}

func NewSaveSnapshotStreamHandler(store domain.SessionStore, repo domain.SnapshotRepository) *SaveSnapshotStreamHandler {
	return &SaveSnapshotStreamHandler{store: store, repo: repo}
}

func (s SaveSnapshotStreamHandler) Handle(ctx context.Context, it domain.UploadIterator) (string, error) {
	var (
		snapshotID   string
		headerSeen   bool
		finalizeSeen bool
	)

	for {
		msg, ok, err := it.Next(ctx)
		if err != nil {
			return "", fmt.Errorf("stream read: %w", err)
		}
		if !ok {
			break // EOF
		}

		if snapshotID == "" {
			snapshotID = msg.SnapshotID
			if snapshotID == "" {
				return "", fmt.Errorf("missing snapshot_id")
			}
		}

		if msg.Header != nil {
			if headerSeen {
				return "", domain.ErrDuplicateHeader
			}
			headerSeen = true
			h := msg.Header
			header := domain.SnapshotHeader{
				SnapshotID:           snapshotID,
				TimestampUnix:        h.TimestampUnix,
				ServerHost:           h.ServerHost,
				ServerType:           h.ServerType,
				ExpectedChunks:       h.ExpectedChunks,
				ExpectedTotalSamples: h.ExpectedTotalSamples,
				AgentVersion:         h.AgentVersion,
				Tags:                 append([]string(nil), h.Tags...),
				MaxChunkBytes:        h.MaxChunkBytes,
			}
			_, err := s.store.UpsertHeader(header)
			if err != nil {
				return "", fmt.Errorf("upsert header: %w", err)
			}
			// ensure snapshot existence in persistence layer (idempotent)
			if err := s.repo.EnsureSnapshot(header); err != nil {
				return "", fmt.Errorf("ensure snapshot: %w", err)
			}
			continue
		}

		if msg.Chunk != nil {
			if !headerSeen {
				return "", domain.ErrChunkBeforeHeader
			}
			ch := msg.Chunk
			if ch.ChunkSeq == 0 {
				return "", domain.ErrInvalidChunkSeq
			}
			already, err := s.store.SaveChunk(snapshotID, ch.ChunkSeq, ch.Samples)
			if err != nil {
				return "", fmt.Errorf("save chunk %d: %w", ch.ChunkSeq, err)
			}
			if !already {
				if err := s.repo.SaveSamples(snapshotID, ch.ChunkSeq, ch.Samples); err != nil {
					return "", fmt.Errorf("repo save chunk %d: %w", ch.ChunkSeq, err)
				}
			}
			continue
		}

		if msg.Finalize != nil {
			if !headerSeen {
				return "", domain.ErrMissingHeader
			}
			if finalizeSeen {
				return "", domain.ErrDuplicateFinalize
			}
			finalizeSeen = true
			if err := s.store.Finalize(snapshotID, msg.Finalize.TotalSamples); err != nil {
				return "", fmt.Errorf("finalize: %w", err)
			}
			continue
		}

		return "", fmt.Errorf("empty upload message payload")
	}

	// Decide outcome after EOF
	st, _, err := s.store.GetMissing(snapshotID)
	if err != nil {
		return "", err
	}

	switch st {
	case domain.UploadStatusComplete:
		if err := s.repo.FinalizeSnapshot(snapshotID); err != nil {
			return "", err
		}
		err = s.store.MarkComplete(snapshotID)
		return snapshotID, err

	case domain.UploadStatusReceiving:
		return snapshotID, domain.ErrMissingChunks

	case domain.UploadStatusUnknown:
		return snapshotID, domain.ErrUnkownSnapId

	case domain.UploadStatusExpired:
		return snapshotID, domain.ErrSessionExpired
	}

	return snapshotID, fmt.Errorf("unknown upload stage")
}
