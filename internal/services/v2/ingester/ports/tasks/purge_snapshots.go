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

type PurgeSnapshotsTask struct {
	interval  time.Duration
	retention time.Duration
	logger    *slog.Logger
	app       app.Application
	batchSize int
}

func NewPurgeSnapshotsTask(application app.Application, interval, retention time.Duration, batchSize int, logger *slog.Logger) *PurgeSnapshotsTask {
	return &PurgeSnapshotsTask{
		interval:  interval,
		retention: retention,
		logger:    logger,
		app:       application,
		batchSize: batchSize,
	}
}

var _ runtimes.Task = (*PurgeSnapshotsTask)(nil)

func (c PurgeSnapshotsTask) Name() string {
	return fmt.Sprintf("PurgeSnapshotsTask")
}

func (c PurgeSnapshotsTask) Interval() time.Duration {
	return c.interval
}

func (c PurgeSnapshotsTask) Run(ctx context.Context) error {
	err := c.app.Command.PurgeSnapshots.Handle(ctx, command.PurgeSnapshots{
		Start:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		End:       time.Now().Add(-c.retention),
		BatchSize: c.batchSize,
	})
	if err != nil {
		return fmt.Errorf("purging snapshots: %w", err)
	}
	return nil
}

func (c PurgeSnapshotsTask) Opts() runtimes.TaskOpts {
	return runtimes.TaskOpts{Overlap: runtimes.OverlapOptDrop}
}
