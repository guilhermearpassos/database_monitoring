package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/collector"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/ingestor"
)

type CollectDeadlocksTask struct {
	targetAlias    string
	server         common_domain.ServerMeta
	databases      []string
	interval       time.Duration
	snapshotter    collector.Snapshotter
	ingestorClient ingestor.Client
	logger         *slog.Logger
}

func NewCollectDeadlocksTask(
	targetAlias string,
	server common_domain.ServerMeta,
	databases []string,
	interval time.Duration,
	snapshotter collector.Snapshotter,
	ingestorClient ingestor.Client,
	logger *slog.Logger,
) *CollectDeadlocksTask {
	return &CollectDeadlocksTask{
		targetAlias:    targetAlias,
		server:         server,
		databases:      databases,
		interval:       interval,
		snapshotter:    snapshotter,
		ingestorClient: ingestorClient,
		logger:         logger,
	}
}

var _ runtimes.Task = (*CollectDeadlocksTask)(nil)

func (c CollectDeadlocksTask) Name() string {
	return fmt.Sprintf("CollectDeadlocksTask(%s)", c.targetAlias)
}

func (c CollectDeadlocksTask) Interval() time.Duration {
	return c.interval
}

func (c CollectDeadlocksTask) Run(ctx context.Context) error {
	deadlocks, err := c.snapshotter.CollectDeadlocks(ctx, c.databases)
	if err != nil {
		return fmt.Errorf("collect deadlocks: %w", err)
	}
	if len(deadlocks) == 0 {
		return nil
	}
	if err := c.ingestorClient.SendDeadlocks(ctx, deadlocks, c.server); err != nil {
		return fmt.Errorf("send deadlocks: %w", err)
	}
	return nil
}

func (c CollectDeadlocksTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
