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

type PurgePlansTask struct {
	interval  time.Duration
	batchSize int
	logger    *slog.Logger
	app       app.Application
}

func NewPurgePlansTask(application app.Application, interval time.Duration, batchSize int, logger *slog.Logger) *PurgePlansTask {
	return &PurgePlansTask{
		interval:  interval,
		batchSize: batchSize,
		logger:    logger,
		app:       application,
	}
}

var _ runtimes.Task = (*PurgePlansTask)(nil)

func (c PurgePlansTask) Name() string {
	return fmt.Sprintf("PurgePlansTask")
}

func (c PurgePlansTask) Interval() time.Duration {
	return c.interval
}

func (c PurgePlansTask) Run(ctx context.Context) error {
	err := c.app.Command.PurgeUnboundedPlans.Handle(ctx, command.PurgePlans{
		BatchSize: c.batchSize,
	})
	if err != nil {
		return fmt.Errorf("purging plans: %w", err)
	}
	return nil
}

func (c PurgePlansTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
