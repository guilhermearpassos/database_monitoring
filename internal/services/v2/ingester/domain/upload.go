package domain

import (
	"context"
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

// UploadMessage is a single message in the upload stream (Header OR Chunk OR Finalize).
// Only one of Header, Chunk, or Finalize should be non-nil in a given message.
type UploadMessage struct {
	SnapshotID string
	Header     *SnapshotHeader
	Chunk      *SampleChunk
	Finalize   *Finalize
}

// UploadIterator abstracts over a streaming transport.
// Next returns (msg, true, nil) for a message; (nil, false, nil) for EOF; or (nil, false, err) on error.
type UploadIterator interface {
	Next(ctx context.Context) (*UploadMessage, bool, error)
}
