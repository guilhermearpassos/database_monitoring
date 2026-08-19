package tasks

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/collector"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/ingestor"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"log/slog"
	"time"
)

type CollectSnapshotTask struct {
	targetalias    string
	databases      []string
	interval       time.Duration
	snapshotter    collector.Snapshotter
	ingestorClient ingestor.Client
	logger         *slog.Logger
}

func NewCollectSnapshotTask(targetalias string, databases []string, interval time.Duration, snapshotter collector.Snapshotter, ingestorClient ingestor.Client, logger *slog.Logger) *CollectSnapshotTask {
	return &CollectSnapshotTask{
		targetalias:    targetalias,
		databases:      databases,
		interval:       interval,
		snapshotter:    snapshotter,
		ingestorClient: ingestorClient,
		logger:         logger,
	}
}

var _ runtimes.Task = (*CollectSnapshotTask)(nil)

func (c CollectSnapshotTask) Name() string {
	return fmt.Sprintf("CollectSnapshotTask(%s)", c.targetalias)
}

func (c CollectSnapshotTask) Interval() time.Duration {
	return c.interval
}

func (c CollectSnapshotTask) Run(ctx context.Context) error {
	snap, err := c.snapshotter.TakeSnapshot(ctx, c.databases)
	if err != nil {
		return fmt.Errorf("collect snapshot: %w", err)
	}
	resp, err := c.ingestorClient.SendSnapshot(ctx, snap, ingestor.SendOptions{
		MaxUncompressedBytes: 1024 * 1024 * 5, // 5MB
		Compression:          ingestorv2.Compression_COMPRESSION_GZIP,
		AgentVersion:         "2",
		Tags:                 nil,
	})
	if err != nil {
		return fmt.Errorf("send snapshot: %w", err)
	}
	if resp.Status != ingestorv2.SnapshotUploadResult_OK {
		c.logger.WarnContext(ctx, "Snapshot upload failed with status %s and message %s", resp.Status, resp.Message)
	}
	return nil
}

func (c CollectSnapshotTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptEnqueue}
}
