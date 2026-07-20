package bootstrap

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/service"
	"log/slog"
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
	AgentConfig service.AgentConfig `toml:"agent" yaml:"agent"`
}
type InfraConfig struct {
	Telemetry telemetry.TelemetryConfig `toml:"telemetry" yaml:"telemetry"`
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
	grpcRuntime, err = runtimes.NewGRPCServerRuntime(cfg.Runtimes.GRPCCfg)
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
	if cfg.Services.AgentConfig.Enabled {
		svc, err := cfg.Services.AgentConfig.GetService(ctx)
		if err != nil {
			panic(err)
		}
		err = svc.Register(grpcRuntime, taskRuntime)
		if err != nil {
			panic(err)
		}
		services = append(services, &svc)
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
