package runtimes

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"time"
)

type GRPCServerRuntime struct {
	lis    net.Listener
	server *grpc.Server
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
