package runtimes

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServerRuntime struct {
	lis    net.Listener
	server *grpc.Server
	inproc *inprocgrpc.Channel
}

func NewGRPCServerRuntime(cfg config.GRPCServerConfig, inproc *inprocgrpc.Channel) (*GRPCServerRuntime, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	server := telemetry.NewGrpcServer(cfg.Grpc.GrpcMessageMaxSize, cfg.Grpc.TLS.Enabled, cfg.Grpc.TLS.CertFile, cfg.Grpc.TLS.KeyFile)

	reflection.Register(server)
	lis, err := net.Listen("tcp", cfg.Grpc.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", cfg.Grpc.Url, err)
	}
	return &GRPCServerRuntime{server: server, lis: lis, inproc: inproc}, nil
}

var _ Runtime = (*GRPCServerRuntime)(nil)

func (r *GRPCServerRuntime) RegisterService(desc *grpc.ServiceDesc, ss any) {
	if r == nil {
		return
	}
	if r.server != nil {
		r.server.RegisterService(desc, ss)
	}
	if r.inproc != nil {
		r.inproc.RegisterService(desc, ss)
	}
}

func (r *GRPCServerRuntime) Type() RuntimeType {
	return GRPCRuntime
}

func (r *GRPCServerRuntime) Start(ctx context.Context) error {
	if r == nil || r.server == nil || r.lis == nil {
		return nil
	}
	erChan := make(chan error, 1)
	go func() {
		if err := r.server.Serve(r.lis); err != nil {
			select {
			case erChan <- err:
			default:
				slog.Error("grpc server stopped with error", "error", err)
			}
		}
	}()
	timer := time.NewTimer(150 * time.Millisecond)
	defer timer.Stop()
	select {
	case err := <-erChan:
		return err
	case <-ctx.Done():
		r.server.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r *GRPCServerRuntime) Stop(ctx context.Context) error {
	if r == nil || r.server == nil {
		return nil
	}
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
		return ctx.Err()
	}
}
