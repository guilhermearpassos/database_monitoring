package runtimes

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type RuntimeType string

const (
	GRPCRuntime    RuntimeType = "GRPC"
	GRPCUIRuntime  RuntimeType = "GRPCUI"
	GRPCWEBRuntime RuntimeType = "GRPCWeb"
	BackGroundTask RuntimeType = "BackGroundTask"
)

type Runtime interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Type() RuntimeType
}
type RuntimeManager struct {
	runtimes map[RuntimeType]Runtime
}

func NewRuntimeManager(runtimes map[RuntimeType]Runtime) *RuntimeManager {
	return &RuntimeManager{runtimes: runtimes}
}

func (rm *RuntimeManager) StartRuntimes(ctx context.Context) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	grpc := rm.runtimes[GRPCRuntime]
	err := grpc.Start(ctx)
	if err != nil {
		return fmt.Errorf("starting grpc runtimes: %w", err)
	}
	for name, rt := range rm.runtimes {
		if name == GRPCRuntime {
			continue
		}
		log.Info(fmt.Sprintf("Starting runtime %s", name))
		err := rt.Start(ctx)
		if err != nil {
			return fmt.Errorf("starting runtime %s: %w", name, err)
		}
		log.Info(fmt.Sprintf("runtime %s started", name))
	}
	return nil
}

func (rm *RuntimeManager) StopRuntimes(ctx context.Context) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	var errs []string
	for name, rt := range rm.runtimes {
		log.Info(fmt.Sprintf("Stopping runtime %s", name))
		err := rt.Stop(ctx)
		if err != nil {
			errs = append(errs, err.Error())
			log.Error(fmt.Sprintf("stopping runtime %s: %s", name, err))
			continue
		}
		log.Info(fmt.Sprintf("runtime %s stopped", name))

	}
	return nil
}

func (rm *RuntimeManager) RegisterRuntime(name RuntimeType, instance Runtime) {
	rm.runtimes[name] = instance
}
func (rm *RuntimeManager) Get(name RuntimeType) Runtime {
	return rm.runtimes[name]
}
