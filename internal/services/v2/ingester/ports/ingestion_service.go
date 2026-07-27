package ports

import (
	"context"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// Compression is an optional transport hint mirrored from the client header.
// It is included here only for observability and future policies. The server
// port does not act on it directly.
// Keep values aligned with the proto if you plan to map them directly.

type Compression int32

const (
	CompressionNone Compression = 0
	CompressionGZIP Compression = 1
	CompressionZSTD Compression = 2
)

// SnapshotHeader represents the header of a single snapshot upload stream.
// A transport adapter (e.g., gRPC) should map incoming proto headers to this DTO.

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
	Compression          Compression // hint only
}

// SampleChunk contains a strictly increasing chunk sequence and the decoded
// samples belonging to that chunk.

type SampleChunk struct {
	ChunkSeq uint32
	Samples  []*dbmv1.QuerySample
}

// Finalize represents the end-of-stream marker with optional totals.

type Finalize struct {
	TotalSamples uint64
}

// UploadMessage is a single message in the upload stream (Header OR Chunk OR Finalize).

type UploadMessage struct {
	SnapshotID string
	Header     *SnapshotHeader
	Chunk      *SampleChunk
	Finalize   *Finalize
}

// UploadIterator abstracts over a streaming transport. It yields messages until
// EOF (ok=false). Returning a non-nil error aborts processing.

type UploadIterator interface {
	Next(ctx context.Context) (*UploadMessage, bool, error)
}

// UploadStatus describes the server-side knowledge of a snapshot upload.

type UploadStatus int32

const (
	UploadStatusUnknown   UploadStatus = 0
	UploadStatusReceiving UploadStatus = 1
	UploadStatusComplete  UploadStatus = 2
	UploadStatusExpired   UploadStatus = 3
)

// SnapshotUploadResultStatus mirrors the high-level outcome of IngestSnapshotStream.

type SnapshotUploadResultStatus int32

const (
	ResultOK       SnapshotUploadResultStatus = 0
	ResultPartial  SnapshotUploadResultStatus = 1
	ResultRejected SnapshotUploadResultStatus = 2
)

// SnapshotUploadResult is returned when a stream finishes (EOF or error).

type SnapshotUploadResult struct {
	Status  SnapshotUploadResultStatus
	Message string
}

// GetMissingChunksResult returns the server awareness and any missing sequences.

type GetMissingChunksResult struct {
	Status        UploadStatus
	MissingChunks []uint32 // valid only when Status == UploadStatusReceiving
}
