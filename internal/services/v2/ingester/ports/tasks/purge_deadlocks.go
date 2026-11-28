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

type PurgeDeadlocksTask struct {
	interval  time.Duration
	retention time.Duration
	logger    *slog.Logger
	app       app.Application
	batchSize int
}

func NewPurgeDeadlocksTask(application app.Application, interval, retention time.Duration, batchSize int, logger *slog.Logger) *PurgeDeadlocksTask {
	return &PurgeDeadlocksTask{
		interval:  interval,
		retention: retention,
		logger:    logger,
		app:       application,
		batchSize: batchSize,
	}
}

var _ runtimes.Task = (*PurgeDeadlocksTask)(nil)

func (c PurgeDeadlocksTask) Name() string {
	return "PurgeDeadlocksTask"
}

func (c PurgeDeadlocksTask) Interval() time.Duration {
	return c.interval
}

func (c PurgeDeadlocksTask) Run(ctx context.Context) error {
	err := c.app.Command.PurgeDeadlocks.Handle(ctx, command.PurgeDeadlocks{
		Start:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:       time.Now().Add(-c.retention),
		BatchSize: c.batchSize,
	})
	if err != nil {
		return fmt.Errorf("purging deadlocks: %w", err)
	}
	return nil
}

func (c PurgeDeadlocksTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
