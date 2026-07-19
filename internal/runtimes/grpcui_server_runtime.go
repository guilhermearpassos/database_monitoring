package runtimes

import (
	"context"
	"fmt"
	grpcui "github.com/fullstorydev/grpcui/standalone"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"net"
	"net/http"
)

type GRPCUiRuntime struct {
	grpcCfg config.GrpcConfig
	cfg     config.GRPCUIConfig
	server  *http.Server
}

func NewGRPCUiRuntime(cfg config.GRPCUIConfig, grpcCfg config.GrpcConfig) (*GRPCUiRuntime, error) {
	return &GRPCUiRuntime{cfg: cfg, grpcCfg: grpcCfg}, nil
}

var _ Runtime = (*GRPCUiRuntime)(nil)

func (G GRPCUiRuntime) Start(ctx context.Context) error {
	if !G.cfg.Enabled {
		return nil
	}
	cc, err := telemetry.OpenInstrumentedClientConn(G.grpcCfg.Url, G.grpcCfg.GrpcMessageMaxSize, G.grpcCfg.TLS.Enabled)
	h, err := grpcui.HandlerViaReflection(ctx, cc, "sqlsights")
	if err != nil {
		return fmt.Errorf("creating grpcui handler: %w", err)
	}
	m := http.NewServeMux()
	lis, err := net.Listen("tcp", G.cfg.Url)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", G.cfg.Url, err)
	}
	go func() {
		m.Handle("/", h)
		server := &http.Server{Handler: m}
		G.server = server
		err = server.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (G GRPCUiRuntime) Stop(ctx context.Context) error {
	if G.server == nil {
		return nil
	}
	return G.server.Shutdown(ctx)
}

func (G GRPCUiRuntime) Type() RuntimeType {
	return GRPCUIRuntime
}
