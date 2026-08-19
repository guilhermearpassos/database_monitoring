package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	agentsvc "github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/service"
	ingestersvc "github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/service"
	queriersvc "github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/service"
	"github.com/guilhermearpassos/database-monitoring/sql"
)

type ApplicationInstance struct {
	Services []*appcommon.Service
	Manager  *runtimes.RuntimeManager
}
type RuntimeCfg struct {
	GRPCCfg   config.GRPCServerConfig `toml:"grpc_server" yaml:"grpc_server"`
	GRPCUICfg config.GRPCUIConfig     `toml:"grpc_ui" yaml:"grpc_ui"`
}
type ServiceCfg struct {
	AgentConfig    agentsvc.AgentConfig       `toml:"agent" yaml:"agent"`
	IngesterConfig ingestersvc.IngesterConfig `toml:"ingester" yaml:"ingester"`
	QuerierConfig  queriersvc.QuerierConfig   `toml:"querier" yaml:"querier"`
}
type InfraConfig struct {
	Telemetry telemetry.TelemetryConfig `toml:"telemetry" yaml:"telemetry"`
	Migrate   MigrateConfig             `toml:"migrate" yaml:"migrate"`
}

type MigrateConfig struct {
	Enabled    bool   `toml:"enabled" yaml:"enabled"`
	ConnString string `toml:"conn_string" yaml:"conn_string"`
}

func (m *MigrateConfig) Migrate() error {
	slog.Info("migrate called")
	d, err := iofs.New(sql.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate fs: %w", err)
	}
	mig, err := migrate.NewWithSourceInstance("iofs", d, m.ConnString)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	err = mig.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	slog.Info("migrate up completed")
	return nil
}

type AppInstanceConfig struct {
	Infra    InfraConfig `toml:"infra" yaml:"infra"`
	Runtimes RuntimeCfg  `toml:"runtimes" yaml:"runtimes"`
	Services ServiceCfg  `toml:"services" yaml:"services"`
}

func NewApplicationInstance(ctx context.Context, cfg AppInstanceConfig) ApplicationInstance {

	err := telemetry.InitTelemetryFromConfig(cfg.Infra.Telemetry)
	if err != nil {
		panic(err)
	}
	var grpcRuntime *runtimes.GRPCServerRuntime
	var taskRuntime *runtimes.BackGroundTaskRuntime
	var grpcUIRuntime *runtimes.GRPCUiRuntime
	inproc := &inprocgrpc.Channel{}
	grpcRuntime, err = runtimes.NewGRPCServerRuntime(cfg.Runtimes.GRPCCfg, inproc)
	if err != nil {
		panic(err)
	}
	grpcUIRuntime, err = runtimes.NewGRPCUiRuntime(cfg.Runtimes.GRPCUICfg, cfg.Runtimes.GRPCCfg.Grpc)
	if err != nil {
		panic(err)
	}
	taskLogger := slog.Default() //TODO improve logging
	taskRuntime = runtimes.NewBackGroundTaskRuntime(taskLogger)
	services := make([]*appcommon.Service, 0)
	if cfg.Infra.Migrate.Enabled {
		err = cfg.Infra.Migrate.Migrate()
		if err != nil {
			panic(err)
		}
	}
	if cfg.Services.AgentConfig.Enabled {
		svc, err := cfg.Services.AgentConfig.GetService(ctx, inproc)
		if err != nil {
			panic(err)
		}
		err = svc.Register(grpcRuntime, taskRuntime)
		if err != nil {
			panic(err)
		}
		services = append(services, &svc)
	}
	if cfg.Services.IngesterConfig.Enabled {
		ingSvc, err := cfg.Services.IngesterConfig.GetService(ctx, inproc)
		if err != nil {
			panic(err)
		}
		err = ingSvc.Register(grpcRuntime, taskRuntime)
		if err != nil {
			panic(err)
		}
		services = append(services, &ingSvc)
	}
	if cfg.Services.QuerierConfig.Enabled {
		querierSvc, err := cfg.Services.QuerierConfig.GetService(ctx, inproc)
		if err != nil {
			panic(err)
		}
		err = querierSvc.Register(grpcRuntime, taskRuntime)
		if err != nil {
			panic(err)
		}
		services = append(services, &querierSvc)
	}
	return ApplicationInstance{
		Services: services,
		Manager: runtimes.NewRuntimeManager(map[runtimes.RuntimeType]runtimes.Runtime{
			runtimes.GRPCUIRuntime:  grpcUIRuntime,
			runtimes.GRPCRuntime:    grpcRuntime,
			runtimes.BackGroundTask: taskRuntime,
		}),
	}
}
func (a *ApplicationInstance) Start(ctx context.Context) error {
	err := a.Manager.StartRuntimes(ctx)
	if err != nil {
		return fmt.Errorf("starting runtimes: %w", err)
	}
	return nil
}
func (a *ApplicationInstance) Stop(ctx context.Context) error {
	err := a.Manager.StopRuntimes(ctx)
	if err != nil {
		return fmt.Errorf("stopping runtimes: %w", err)
	}
	return nil
}
