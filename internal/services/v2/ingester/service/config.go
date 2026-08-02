package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/adapters/repository"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/adapters/state"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/ports"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/ports/tasks"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
)

// IngesterConfig mirrors the style of the v2 agent service configuration.
// It provides minimal knobs for the ingestion server application layer.
// Adapter wiring (e.g., gRPC) is intentionally left out for now.
type IngesterConfig struct {
	Enabled   bool                  `yaml:"enabled" toml:"enabled"`
	Postgres  config.PostgresConfig `yaml:"postgres" toml:"postgres"`
	Retention struct {
		Snapshots RetentionConfig `yaml:"snapshots" toml:"snapshots"`
		Metrics   RetentionConfig `yaml:"metrics" toml:"metrics"`
	} `yaml:"retention" toml:"retention"`
}

type RetentionConfig struct {
	Retention time.Duration `yaml:"retention" toml:"retention"`
	Interval  time.Duration `yaml:"interval" toml:"interval"`
}

func (r RetentionConfig) GetInterval() time.Duration {
	if r.Interval == 0 {
		return 10 * time.Minute
	}
	return r.Interval
}
func (r RetentionConfig) GetRetention() time.Duration {
	if r.Retention == 0 {
		return 240 * time.Hour
	}
	return r.Retention
}

// GetService constructs the transport-agnostic ports.GrpcIngester using
// in-memory/default adapters so the service can run without external deps.
func (c *IngesterConfig) GetService(ctx context.Context, inproc *inprocgrpc.Channel) (appcommon.Service, error) { //nolint:revive,unused
	store := state.NewMemoryStore()
	var repo interface {
		domain.SnapshotRepository
		domain.QueryMetricsRepository
	}
	if c.Postgres.Connstring != "" {
		db, err := c.Postgres.Get(ctx)
		if err != nil {
			return nil, err
		}
		repo = repository.NewPostgresRepo(db)
	}
	if repo == nil {
		return nil, fmt.Errorf("no valid repo config found")
	}
	application := app.NewApplicationWithAdapters(store, repo, repo)
	p := ports.NewService(application)
	tsks := []runtimes.Task{
		tasks.NewPurgeQueryMetricsTask(application, c.Retention.Metrics.GetInterval(), c.Retention.Metrics.GetRetention(), slog.Default()),
		tasks.NewPurgeSnapshotsTask(application, c.Retention.Snapshots.GetInterval(), c.Retention.Snapshots.GetRetention(), slog.Default()),
	}
	return IngesterService{Port: p, Tasks: tsks}, nil
}

// IngesterService conforms to appcommon.Service and is the entry point the
// bootstrap will register. For now, it does not bind any transport.
type IngesterService struct {
	Port  *ports.GrpcIngester
	Tasks []runtimes.Task
}

func (s IngesterService) Register(grpcRuntime *runtimes.GRPCServerRuntime, TaskRuntime *runtimes.BackGroundTaskRuntime) error {
	// No transport adapter yet; nothing to register.
	// The Port is ready to be bound by a future gRPC adapter and/or inproc hub.
	grpcRuntime.RegisterService(&ingestorv2.IngestionService_ServiceDesc, s.Port)
	for _, task := range s.Tasks {
		err := TaskRuntime.RegisterTask(task)
		if err != nil {
			return fmt.Errorf("register task %s error: %w", task.Name(), err)
		}
	}
	return nil
}

var _ appcommon.Service = (*IngesterService)(nil)
