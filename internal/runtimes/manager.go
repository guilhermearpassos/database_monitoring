package runtimes

import (
	"context"
	"errors"
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
	var started []Runtime
	rollback := func() {
		for i := len(started) - 1; i >= 0; i-- {
			_ = started[i].Stop(ctx)
		}
	}
	if grpc, ok := rm.runtimes[GRPCRuntime]; ok && grpc != nil {
		if err := grpc.Start(ctx); err != nil {
			return fmt.Errorf("starting grpc runtimes: %w", err)
		}
		started = append(started, grpc)
	}
	for name, rt := range rm.runtimes {
		if name == GRPCRuntime || rt == nil {
			continue
		}
		log.Info(fmt.Sprintf("Starting runtime %s", name))
		err := rt.Start(ctx)
		if err != nil {
			rollback()
			return fmt.Errorf("starting runtime %s: %w", name, err)
		}
		started = append(started, rt)
		log.Info(fmt.Sprintf("runtime %s started", name))
	}
	return nil
}

func (rm *RuntimeManager) StopRuntimes(ctx context.Context) error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// Stop dependent runtimes first, then gRPC server last.
	order := make([]RuntimeType, 0, len(rm.runtimes))
	for name := range rm.runtimes {
		if name != GRPCRuntime {
			order = append(order, name)
		}
	}
	if _, ok := rm.runtimes[GRPCRuntime]; ok {
		order = append(order, GRPCRuntime)
	}

	var errs error
	for _, name := range order {
		rt := rm.runtimes[name]
		if rt == nil {
			continue
		}
		log.Info(fmt.Sprintf("Stopping runtime %s", name))
		err := rt.Stop(ctx)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("stopping runtime %s: %w", name, err))
			log.Error(fmt.Sprintf("stopping runtime %s: %s", name, err))
			continue
		}
		log.Info(fmt.Sprintf("runtime %s stopped", name))
	}
	return errs
}

func (rm *RuntimeManager) RegisterRuntime(name RuntimeType, instance Runtime) {
	rm.runtimes[name] = instance
}
func (rm *RuntimeManager) Get(name RuntimeType) Runtime {
	return rm.runtimes[name]
}
