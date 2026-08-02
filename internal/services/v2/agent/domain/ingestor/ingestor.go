package ingestor

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
)

// SendOptions controls per-upload behavior for the v2 ingestion client.
// This lives in the domain layer so adapters can implement the same contract.
type SendOptions struct {
	// MaxUncompressedBytes caps the size of a SampleChunk message before transport compression.
	// The client will pack as many samples as fit under this limit.
	MaxUncompressedBytes int
	// Compression hint propagated in header (server may ignore).
	Compression ingestorv2.Compression
	// AgentVersion and Tags are attached to header for observability.
	AgentVersion string
	Tags         []string
}

// Client defines the domain contract for sending a snapshot to the ingester.
// Adapters (e.g., gRPC client) should implement this interface.
type Client interface {
	SendSnapshot(ctx context.Context, snap *common_domain.DataBaseSnapshot, so SendOptions) (*ingestorv2.SnapshotUploadResult, error)
	SendExecutionPlans(ctx context.Context, plans []*common_domain.ExecutionPlan, server common_domain.ServerMeta, so SendOptions) (*ingestorv2.ExecutionPlanUploadResult, error)
}
