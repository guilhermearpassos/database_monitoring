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

type CollectQueryMetricsTask struct {
	targetAlias    string
	server         common_domain.ServerMeta
	databases      []string
	interval       time.Duration
	snapshotter    collector.Snapshotter
	ingestorClient ingestor.Client
	logger         *slog.Logger
}

func NewCollectQueryMetricsTask(
	targetAlias string,
	server common_domain.ServerMeta,
	databases []string,
	interval time.Duration,
	snapshotter collector.Snapshotter,
	ingestorClient ingestor.Client,
	logger *slog.Logger,
) *CollectQueryMetricsTask {
	return &CollectQueryMetricsTask{
		targetAlias:    targetAlias,
		server:         server,
		databases:      databases,
		interval:       interval,
		snapshotter:    snapshotter,
		ingestorClient: ingestorClient,
		logger:         logger,
	}
}

var _ runtimes.Task = (*CollectQueryMetricsTask)(nil)

func (c CollectQueryMetricsTask) Name() string {
	return fmt.Sprintf("CollectQueryMetricsTask(%s)", c.targetAlias)
}

func (c CollectQueryMetricsTask) Interval() time.Duration {
	return c.interval
}

func (c CollectQueryMetricsTask) Run(ctx context.Context) error {
	collectedAt := time.Now().UTC()
	metrics, err := c.snapshotter.CollectQueryMetrics(ctx, c.databases)
	if err != nil {
		return fmt.Errorf("collect query metrics: %w", err)
	}
	if len(metrics) == 0 {
		return nil
	}
	if err := c.ingestorClient.SendQueryMetrics(ctx, metrics, c.server, collectedAt); err != nil {
		return fmt.Errorf("send query metrics: %w", err)
	}
	return nil
}

func (c CollectQueryMetricsTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
