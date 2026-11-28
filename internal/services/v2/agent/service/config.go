package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/adapters/collector"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/adapters/ingestor"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/ports/tasks"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
	"google.golang.org/grpc"
)

type TargetConfig struct {
	Alias               string              `yaml:"alias" toml:"alias"`
	Driver              string              `yaml:"driver" toml:"driver"`
	ConnString          string              `yaml:"conn_string" toml:"conn_string"`
	SnapInterval        time.Duration       `yaml:"snap_interval" toml:"snap_interval"`
	Interval            time.Duration       `yaml:"interval" toml:"interval"`
	PlanCollection      PlanCollection      `yaml:"plan" toml:"plan"`
	MetricsCollection   MetricsCollection   `yaml:"metrics" toml:"metrics"`
	DeadlocksCollection DeadlocksCollection `yaml:"deadlocks" toml:"deadlocks"`
}

func (cfg *TargetConfig) GetSnapInterval() time.Duration {
	if cfg.SnapInterval > 0 {
		return cfg.SnapInterval
	}
	if cfg.Interval > 0 {
		return cfg.Interval
	}
	return 10 * time.Second
}

type PlanCollection struct {
	Disabled bool          `yaml:"disabled" toml:"disabled"`
	Interval time.Duration `yaml:"interval" toml:"interval"`
	Lookback time.Duration `yaml:"lookback" toml:"lookback"`
}

type MetricsCollection struct {
	Disabled bool          `yaml:"disabled" toml:"disabled"`
	Interval time.Duration `yaml:"interval" toml:"interval"`
}

func (cfg *MetricsCollection) GetMetricsInterval() time.Duration {
	if cfg.Interval == 0 {
		return 30 * time.Second
	}
	return cfg.Interval
}

type DeadlocksCollection struct {
	Disabled bool          `yaml:"disabled" toml:"disabled"`
	Interval time.Duration `yaml:"interval" toml:"interval"`
}

func (cfg *DeadlocksCollection) GetDeadlocksInterval() time.Duration {
	if cfg.Interval == 0 {
		return 30 * time.Second
	}
	return cfg.Interval
}

func (cfg *PlanCollection) GetPlanInterval() time.Duration {
	if cfg.Interval == 0 {
		return 10 * time.Second
	}
	return cfg.Interval
}
func (cfg *PlanCollection) GetPlanLookback() time.Duration {
	if cfg.Lookback == 0 {
		return 3 * cfg.GetPlanInterval()
	}
	return cfg.Lookback
}

type AgentConfig struct {
	Targets        []TargetConfig             `yaml:"targets" toml:"targets"`
	IngestorClient telemetry.GRPCClientConfig `yaml:"ingestor_client" toml:"ingestor_client"`
	Enabled        bool                       `yaml:"enabled" toml:"enabled"`
}

func (c *AgentConfig) GetService(ctx context.Context, inproc *inprocgrpc.Channel) (appcommon.Service, error) {
	tsks := make(map[string]runtimes.Task, len(c.Targets))
	logger := slog.Default()

	var conn grpc.ClientConnInterface
	if c.IngestorClient.Endpoint == "" || strings.HasPrefix(c.IngestorClient.Endpoint, "inproc://") {
		conn = inproc
	} else {
		maxMsgSize := c.IngestorClient.MaxMsgSize
		if maxMsgSize <= 0 {
			maxMsgSize = 100 * 1024 * 1024
		}
		cc, err := telemetry.OpenInstrumentedClientConnWithTLS(c.IngestorClient.Endpoint, maxMsgSize, c.IngestorClient.TlsEnabled, c.IngestorClient.TlsInsecureSkipVerify)
		if err != nil {
			return nil, fmt.Errorf("dial ingestor %s: %w", c.IngestorClient.Endpoint, err)
		}
		conn = cc
	}
	isc := ingestorv2.NewIngestionServiceClient(conn)
	ic, err := ingestor.New(isc)
	if err != nil {
		return nil, err
	}

	for _, target := range c.Targets {
		server := common_domain.ServerMeta{
			Host: target.Alias,
			Type: target.Driver,
		}
		db, err := telemetry.OpenInstrumentedDB(target.Driver, target.ConnString)
		if err != nil {
			return nil, fmt.Errorf("open instrumented DB: %w", err)
		}
		ss := collector.NewSqlServerSnapshotter(db, server)

		t := tasks.NewCollectSnapshotTask(target.Alias, []string{}, target.GetSnapInterval(), ss, ic, logger)
		tsks[t.Name()] = t
		planCfg := target.PlanCollection
		if !planCfg.Disabled {
			t2 := tasks.NewCollectExecutionPlansTask(target.Alias, server, []string{}, planCfg.GetPlanInterval(), planCfg.GetPlanLookback(), ss, ic, logger)
			tsks[t2.Name()] = t2
		}
		metricsCfg := target.MetricsCollection
		if !metricsCfg.Disabled {
			t3 := tasks.NewCollectQueryMetricsTask(target.Alias, server, []string{}, metricsCfg.GetMetricsInterval(), ss, ic, logger)
			tsks[t3.Name()] = t3
		}
		deadlocksCfg := target.DeadlocksCollection
		if !deadlocksCfg.Disabled {
			t4 := tasks.NewCollectDeadlocksTask(target.Alias, server, []string{}, deadlocksCfg.GetDeadlocksInterval(), ss, ic, logger)
			tsks[t4.Name()] = t4
		}
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
