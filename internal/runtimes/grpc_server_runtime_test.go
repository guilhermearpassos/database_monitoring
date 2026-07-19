package runtimes

import (
    "context"
    "net"
    "testing"
    "time"

    "google.golang.org/grpc"
)

func TestGRPCServerRuntime_StartStop(t *testing.T) {
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        t.Fatalf("listen: %v", err)
    }
    srv := grpc.NewServer()
    rt := &GRPCServerRuntime{lis: lis, server: srv}

    if err := rt.Start(context.Background()); err != nil {
        t.Fatalf("start: %v", err)
    }

    stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    if err := rt.Stop(stopCtx); err != nil {
        t.Fatalf("stop: %v", err)
    }
}
