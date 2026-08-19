package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app/command"
)

type PurgeQueryMetricsTask struct {
	interval  time.Duration
	retention time.Duration
	logger    *slog.Logger
	app       app.Application
	batchSize int
}

func NewPurgeQueryMetricsTask(application app.Application, interval, retention time.Duration, batchSize int, logger *slog.Logger) *PurgeQueryMetricsTask {
	return &PurgeQueryMetricsTask{
		interval:  interval,
		retention: retention,
		logger:    logger,
		app:       application,
		batchSize: batchSize,
	}
}

var _ runtimes.Task = (*PurgeQueryMetricsTask)(nil)

func (c PurgeQueryMetricsTask) Name() string {
	return fmt.Sprintf("PurgeQueryMetricsTask")
}

func (c PurgeQueryMetricsTask) Interval() time.Duration {
	return c.interval
}

func (c PurgeQueryMetricsTask) Run(ctx context.Context) error {
	err := c.app.Command.PurgeQueryMetrics.Handle(ctx, command.PurgeQueryMetrics{
		Start:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:       time.Now().Add(-c.retention),
		BatchSize: c.batchSize,
	})
	if err != nil {
		return fmt.Errorf("purging query metrics: %w", err)
	}
	return nil
}

func (c PurgeQueryMetricsTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
