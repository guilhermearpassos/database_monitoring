package domain

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// SampleChunk contains a strictly increasing chunk sequence and its decoded samples.
// This is a domain-level type independent from transport/protobuf contracts.
// The concrete protobuf message is adapted to this type in the ports layer.
type SampleChunk struct {
	ChunkSeq uint32
	Samples  []*dbmv1.QuerySample
}

// Finalize represents the end-of-stream marker with optional totals.
type Finalize struct {
	TotalSamples uint64
}

// SnapUploadMessage is a single message in the upload stream (Header OR Chunk OR Finalize).
// Only one of Header, Chunk, or Finalize should be non-nil in a given message.
type SnapUploadMessage struct {
	SnapshotID string
	Header     *SnapshotHeader
	Chunk      *SampleChunk
	Finalize   *Finalize
}
type PlanUploadMessage struct {
	Header   *PlanHeader
	Chunk    *PlanChunk
	Finalize *Finalize
}

// SnapshotHeader holds metadata for a snapshot upload session.
type PlanHeader struct {
	TimestampUnix      int64
	ServerHost         string
	ServerType         string
	ExpectedChunks     uint32
	ExpectedTotalPlans uint64
	AgentVersion       string
	Tags               []string
	MaxChunkBytes      uint32
}
type PlanChunk struct {
	ChunkSeq       uint32
	ExecutionPlans []*common_domain.ExecutionPlan
}

// UploadIterator abstracts over a streaming transport.
// Next returns (msg, nil) for a message; (nil, nil) for EOF; or (nil, err) on error.
type UploadIterator[T any] interface {
	Next(ctx context.Context) (*T, error)
}
