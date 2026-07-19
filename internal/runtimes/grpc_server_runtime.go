package runtimes

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type GRPCServerRuntime struct {
	lis    net.Listener
	server *grpc.Server
}

func NewGRPCServerRuntime(cfg config.GRPCServerConfig) (*GRPCServerRuntime, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	server := telemetry.NewGrpcServer(cfg.Grpc.GrpcMessageMaxSize, cfg.Grpc.TLS.Enabled, cfg.Grpc.TLS.CertFile, cfg.Grpc.TLS.KeyFile)
	lis, err := net.Listen("tcp", cfg.Grpc.Url)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", cfg.Grpc.Url, err)
	}
	return &GRPCServerRuntime{server: server, lis: lis}, nil
}

var _ Runtime = (*GRPCServerRuntime)(nil)

func (r *GRPCServerRuntime) Type() RuntimeType {
	return GRPCRuntime
}

func (r *GRPCServerRuntime) Start(ctx context.Context) error {
	erChan := make(chan error, 1)
	go func() {
		if err := r.server.Serve(r.lis); err != nil {
			select {
			case erChan <- err:
			default:
				panic(fmt.Sprintf("grpc server failed to serve: %v", err))
			}
		}
	}()
	select {
	case err := <-erChan:
		return err
	case <-time.After(1 * time.Second):
		return nil
	}
}

func (r *GRPCServerRuntime) Stop(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		r.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		r.server.Stop()
		return nil
	}
}
