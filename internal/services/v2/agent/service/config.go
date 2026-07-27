package service

import (
	"context"
	"fmt"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/adapters/collector"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/adapters/ingestor"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/ports/tasks"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"log/slog"
	"time"
)

type TargetConfig struct {
	Alias      string        `yaml:"alias" toml:"alias"`
	Driver     string        `yaml:"driver" toml:"driver"`
	ConnString string        `yaml:"conn_string" toml:"conn_string"`
	Interval   time.Duration `yaml:"interval" toml:"interval"`
}

func (cfg *TargetConfig) GetInterval() time.Duration {
	if cfg.Interval == 0 {
		return 10 * time.Second
	}
	return cfg.Interval
}

type AgentConfig struct {
	Targets        []TargetConfig             `yaml:"targets" toml:"targets"`
	IngestorClient telemetry.GRPCClientConfig `yaml:"ingestor_client" toml:"ingestor_client"`
	Enabled        bool                       `yaml:"enabled" toml:"enabled"`
}

func (c *AgentConfig) GetService(ctx context.Context, inproc *inprocgrpc.Channel) (appcommon.Service, error) {
	tsks := make(map[string]runtimes.Task, len(c.Targets))
	logger := slog.Default()
	for _, target := range c.Targets {
		db, err := telemetry.OpenInstrumentedDB(target.Driver, target.ConnString)
		if err != nil {
			return nil, fmt.Errorf("open instrumented DB: %w", err)
		}
		ss := collector.NewSqlServerSnapshotter(db)

		isc := ingestorv2.NewIngestionServiceClient(inproc)
		ic, err := ingestor.New(isc)
		if err != nil {
			return nil, err
		}
		t := tasks.NewCollectSnapshotTask(target.Alias, []string{}, target.GetInterval(), ss, ic, logger)
		tsks[t.Name()] = t
	}
	return AgentService{Tasks: tsks}, nil
}

type AgentService struct {
	Tasks map[string]runtimes.Task
}

func (a AgentService) Register(grpcRuntime *runtimes.GRPCServerRuntime, TaskRuntime *runtimes.BackGroundTaskRuntime) error {
	for name, task := range a.Tasks {
		err := TaskRuntime.RegisterTask(task)
		if err != nil {
			return fmt.Errorf("register task %s error: %w", name, err)
		}
	}
	return nil
}

var _ appcommon.Service = (*AgentService)(nil)
