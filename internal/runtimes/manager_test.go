package runtimes

import (
    "context"
    "errors"
    "sync/atomic"
    "testing"
    "time"
)

type fakeRuntime struct {
    t RuntimeType
    starts atomic.Int64
    stops  atomic.Int64
    startErr error
    stopErr  error
}

func (f *fakeRuntime) Start(ctx context.Context) error { f.starts.Add(1); return f.startErr }
func (f *fakeRuntime) Stop(ctx context.Context) error  { f.stops.Add(1); return f.stopErr }
func (f *fakeRuntime) Type() RuntimeType               { return f.t }

func TestRuntimeManager_StartStop_CallsAll(t *testing.T) {
    rm := &RuntimeManager{runtimes: map[RuntimeType]Runtime{}}
    a := &fakeRuntime{t: GRPCRuntime}
    b := &fakeRuntime{t: BackGroundTask}
    rm.RegisterRuntime(a.Type(), a)
    rm.RegisterRuntime(b.Type(), b)

    if err := rm.StartRuntimes(context.Background()); err != nil {
        t.Fatalf("start runtimes: %v", err)
    }
    stopCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    if err := rm.StopRuntimes(stopCtx); err != nil {
        t.Fatalf("stop runtimes: %v", err)
    }

    if a.starts.Load() != 1 || b.starts.Load() != 1 {
        t.Fatalf("expected starts=1; got a=%d b=%d", a.starts.Load(), b.starts.Load())
    }
    if a.stops.Load() != 1 || b.stops.Load() != 1 {
        t.Fatalf("expected stops=1; got a=%d b=%d", a.stops.Load(), b.stops.Load())
    }
}

func TestRuntimeManager_NoGRPCRuntimeAndStopErrorPropagation(t *testing.T) {
    rm := &RuntimeManager{runtimes: map[RuntimeType]Runtime{}}
    b := &fakeRuntime{t: BackGroundTask, stopErr: errors.New("stop boom")}
    rm.RegisterRuntime(b.Type(), b)

    if err := rm.StartRuntimes(context.Background()); err != nil {
        t.Fatalf("expected start without GRPCRuntime to succeed, got: %v", err)
    }
    if err := rm.StopRuntimes(context.Background()); err == nil {
        t.Fatalf("expected StopRuntimes to return joined stop error, got nil")
    }
}
