package bootstrap

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"log/slog"
)

type ApplicationInstance struct {
	Services []*Service
	Manager  *runtimes.RuntimeManager
}
type Service interface { //Agent extracts data
	Regiter(grpcRuntime runtimes.GRPCServerRuntime, TaskRuntime runtimes.BackGroundTaskRuntime) error
}
type RuntimeCfg struct {
	GRPCCfg   config.GRPCServerConfig `toml:"grpc_server"`
	GRPCUICfg config.GRPCUIConfig     `toml:"grpc_ui"`
}
type ServiceCfg struct {
}
type AppInstanceConfig struct {
	Runtimes RuntimeCfg `toml:"runtimes"`
	Services ServiceCfg `toml:"services"`
}

func NewApplicationInstance(cfg AppInstanceConfig) ApplicationInstance {
	var grpcRuntime *runtimes.GRPCServerRuntime
	var taskRuntime *runtimes.BackGroundTaskRuntime
	var grpcUIRuntime *runtimes.GRPCUiRuntime
	var err error
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
	return ApplicationInstance{
		Services: make([]*Service, 0),
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
}
