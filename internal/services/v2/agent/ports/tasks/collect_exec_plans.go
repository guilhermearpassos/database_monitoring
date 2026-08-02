package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/collector"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/ingestor"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"golang.org/x/exp/maps"
)

type CollectExecutionPlansTask struct {
	targetalias    string
	databases      []string
	interval       time.Duration
	lookback       time.Duration
	snapshotter    collector.Snapshotter
	ingestorClient ingestor.Client
	logger         *slog.Logger
}

func NewCollectExecutionPlansTask(targetalias string, databases []string, interval, lookback time.Duration, snapshotter collector.Snapshotter, ingestorClient ingestor.Client, logger *slog.Logger) *CollectExecutionPlansTask {
	return &CollectExecutionPlansTask{
		targetalias:    targetalias,
		databases:      databases,
		interval:       interval,
		lookback:       lookback,
		snapshotter:    snapshotter,
		ingestorClient: ingestorClient,
		logger:         logger,
	}
}

var _ runtimes.Task = (*CollectExecutionPlansTask)(nil)

func (c CollectExecutionPlansTask) Name() string {
	return fmt.Sprintf("CollectExecutionPlansTask(%s)", c.targetalias)
}

func (c CollectExecutionPlansTask) Interval() time.Duration {
	return c.interval
}

func (c CollectExecutionPlansTask) Run(ctx context.Context) error {
	handles := make([]string, 0)

	plans, err := c.snapshotter.FetchExecutionPlans(ctx, handles)
	if plans != nil && len(plans.Plans) > 0 {
		_, err2 := c.ingestorClient.SendExecutionPlans(ctx, maps.Values(plans.Plans), plans.Server, ingestor.SendOptions{
			MaxUncompressedBytes: 1024 * 1024 * 5, // 5MB
			Compression:          ingestorv2.Compression_COMPRESSION_GZIP,
			AgentVersion:         "2",
			Tags:                 nil,
		})
		if err2 != nil {
			return fmt.Errorf("send plans: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("collect plans: %w", err)
	}
	return nil
}

func (c CollectExecutionPlansTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptEnqueue}
}
