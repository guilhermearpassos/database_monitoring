package runtimes

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	grpcui "github.com/fullstorydev/grpcui/standalone"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"google.golang.org/grpc"
)

type GRPCUiRuntime struct {
	grpcCfg config.GrpcConfig
	cfg     config.GRPCUIConfig

	mu     sync.Mutex
	server *http.Server
	cc     *grpc.ClientConn
}

func NewGRPCUiRuntime(cfg config.GRPCUIConfig, grpcCfg config.GrpcConfig) (*GRPCUiRuntime, error) {
	return &GRPCUiRuntime{cfg: cfg, grpcCfg: grpcCfg}, nil
}

var _ Runtime = (*GRPCUiRuntime)(nil)

func (G *GRPCUiRuntime) Start(ctx context.Context) error {
	if G == nil || !G.cfg.Enabled {
		return nil
	}
	cc, err := telemetry.OpenInstrumentedClientConn(G.grpcCfg.Url, G.grpcCfg.GrpcMessageMaxSize, G.grpcCfg.TLS.Enabled)
	if err != nil {
		return fmt.Errorf("dialing grpc for grpcui: %w", err)
	}
	h, err := grpcui.HandlerViaReflection(ctx, cc, "sqlsights")
	if err != nil {
		_ = cc.Close()
		return fmt.Errorf("creating grpcui handler: %w", err)
	}
	m := http.NewServeMux()
	m.Handle("/", h)
	lis, err := net.Listen("tcp", G.cfg.Url)
	if err != nil {
		_ = cc.Close()
		return fmt.Errorf("listening on %s: %w", G.cfg.Url, err)
	}
	server := &http.Server{
		Handler:           m,
		ReadHeaderTimeout: 10 * time.Second,
	}
	G.mu.Lock()
	G.cc = cc
	G.server = server
	G.mu.Unlock()

	go func() {
		if err := server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("grpcui server stopped with error", "error", err)
		}
	}()
	return nil
}

func (G *GRPCUiRuntime) Stop(ctx context.Context) error {
	if G == nil {
		return nil
	}
	G.mu.Lock()
	srv := G.server
	cc := G.cc
	G.server = nil
	G.cc = nil
	G.mu.Unlock()

	var errs error
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if cc != nil {
		if err := cc.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (G *GRPCUiRuntime) Type() RuntimeType {
	return GRPCUIRuntime
}
