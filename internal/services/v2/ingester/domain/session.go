package domain

import (
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"time"
)

// SnapshotHeader holds metadata for a snapshot upload session.
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

// SnapshotSession is the resumable upload session state.
type SnapshotSession struct {
	Header         SnapshotHeader
	Received       map[uint32]struct{} // chunk_seq received
	ExpectedChunks uint32              // convenience copy
	ExpiresAt      time.Time
	Finalized      bool // finalize message was received
	Completed      bool // all expected chunks have been received
}

// UploadStatus enumerates session knowledge.
type UploadStatus int32

const (
	UploadStatusUnknown   UploadStatus = 0
	UploadStatusReceiving UploadStatus = 1
	UploadStatusComplete  UploadStatus = 2
	UploadStatusExpired   UploadStatus = 3
)

// SessionStore abstracts resumable, idempotent state for uploads.
type SessionStore interface {
	// UpsertHeader creates or returns existing session for SnapshotID (idempotent).
	UpsertHeader(h SnapshotHeader) (*SnapshotSession, error)
	// SaveChunk stores a chunk payload, deduping by (snapshotID, seq). Returns true if already had it.
	SaveChunk(snapshotID string, seq uint32, samples []*dbmv1.QuerySample) (already bool, err error)
	// Finalize records finalize intent and can validate totals.
	Finalize(snapshotID string, totalSamples uint64) error
	// GetMissing reports session status and missing sequences.
	GetMissing(snapshotID string) (UploadStatus, []uint32, error)
	// MarkComplete allows store cleanup when fully complete.
	MarkComplete(snapshotID string) error
}
