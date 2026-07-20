package command

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// Config controls small guardrails at the application level.
// You can extend this with size limits, timeouts, etc.
type Config struct {
	HeaderTTL       time.Duration
	MaxChunkSamples int // 0 disables this check
}

type Service struct {
	Store  domain.SessionStore
	Repo   domain.SnapshotRepository
	Config Config
}

// UploadIterator is a transport-agnostic iterator used by the app layer.
// Adapters should provide an implementation that yields header/chunk/finalize messages.
type UploadIterator interface {
	Next(ctx context.Context) (*UploadMessage, bool, error)
}

// App-layer DTOs (no dependency on ports to avoid cycles)

type SnapshotHeader struct {
	SnapshotID           string
	TimestampUnix        int64
	ServerHost           string
	ServerType           string
	ExpectedChunks       uint32
	ExpectedTotalSamples uint64
	AgentVersion         string
	Tags                 []string
	MaxChunkBytes        uint32
}

type SampleChunk struct {
	ChunkSeq uint32
	Samples  []*dbmv1.QuerySample
}

type Finalize struct {
	TotalSamples uint64
}

type UploadMessage struct {
	SnapshotID string
	Header     *SnapshotHeader
	Chunk      *SampleChunk
	Finalize   *Finalize
}

type ResultStatus int32

const (
	ResultOK ResultStatus = iota
	ResultPartial
	ResultRejected
)

type Result struct {
	Status        ResultStatus
	SnapshotID    string
	MissingChunks []uint32
	Message       string
}

// IngestSnapshotStream consumes a transport-agnostic iterator that yields
// Header, Chunk, Finalize messages for a single snapshot stream and orchestrates
// the domain session store and repository to produce a final result.
func (s Service) IngestSnapshotStream(ctx context.Context, it UploadIterator) (*Result, error) {
	var (
		snapshotID   string
		headerSeen   bool
		finalizeSeen bool
	)

	for {
		msg, ok, err := it.Next(ctx)
		if err != nil {
			return nil, fmt.Errorf("stream read: %w", err)
		}
		if !ok {
			break // EOF
		}

		if snapshotID == "" {
			snapshotID = msg.SnapshotID
			if snapshotID == "" {
				return nil, fmt.Errorf("missing snapshot_id")
			}
		}

		if msg.Header != nil {
			if headerSeen {
				return nil, domain.ErrDuplicateHeader
			}
			headerSeen = true
			h := msg.Header
			_, err := s.Store.UpsertHeader(domain.SnapshotHeader{
				SnapshotID:           snapshotID,
				TimestampUnix:        h.TimestampUnix,
				ServerHost:           h.ServerHost,
				ServerType:           h.ServerType,
				ExpectedChunks:       h.ExpectedChunks,
				ExpectedTotalSamples: h.ExpectedTotalSamples,
				AgentVersion:         h.AgentVersion,
				Tags:                 append([]string(nil), h.Tags...),
				MaxChunkBytes:        h.MaxChunkBytes,
			}, s.Config.HeaderTTL)
			if err != nil {
				return nil, fmt.Errorf("upsert header: %w", err)
			}
			continue
		}

		if msg.Chunk != nil {
			if !headerSeen {
				return nil, domain.ErrChunkBeforeHeader
			}
			ch := msg.Chunk
			if ch.ChunkSeq == 0 {
				return nil, domain.ErrInvalidChunkSeq
			}
			if s.Config.MaxChunkSamples > 0 && len(ch.Samples) > s.Config.MaxChunkSamples {
				return nil, fmt.Errorf("too many samples in chunk: %d", len(ch.Samples))
			}

			already, err := s.Store.SaveChunk(snapshotID, ch.ChunkSeq, ch.Samples)
			if err != nil {
				return nil, fmt.Errorf("save chunk %d: %w", ch.ChunkSeq, err)
			}
			if !already {
				if err := s.Repo.SaveSamples(snapshotID, ch.ChunkSeq, ch.Samples); err != nil {
					return nil, fmt.Errorf("repo save chunk %d: %w", ch.ChunkSeq, err)
				}
			}
			continue
		}

		if msg.Finalize != nil {
			if !headerSeen {
				return nil, domain.ErrMissingHeader
			}
			if finalizeSeen {
				return nil, domain.ErrDuplicateFinalize
			}
			finalizeSeen = true
			if err := s.Store.Finalize(snapshotID, msg.Finalize.TotalSamples); err != nil {
				return nil, fmt.Errorf("finalize: %w", err)
			}
			continue
		}

		return nil, fmt.Errorf("empty upload message payload")
	}

	// Decide outcome after EOF
	st, missing, err := s.Store.GetMissing(snapshotID)
	if err != nil {
		return nil, err
	}

	switch st {
	case domain.UploadStatusComplete:
		if err := s.Repo.FinalizeSnapshot(snapshotID); err != nil {
			return nil, err
		}
		_ = s.Store.MarkComplete(snapshotID)
		return &Result{Status: ResultOK, SnapshotID: snapshotID}, nil

	case domain.UploadStatusReceiving:
		return &Result{Status: ResultPartial, SnapshotID: snapshotID, MissingChunks: missing, Message: "awaiting missing chunks"}, nil

	case domain.UploadStatusUnknown:
		return &Result{Status: ResultRejected, SnapshotID: snapshotID, Message: "unknown snapshot_id"}, nil

	case domain.UploadStatusExpired:
		return &Result{Status: ResultRejected, SnapshotID: snapshotID, Message: "session expired"}, nil
	}

	return &Result{Status: ResultRejected, SnapshotID: snapshotID, Message: "unexpected state"}, nil
}
