package runtimes

import (
    "context"
    "io"
    "log/slog"
    "sync/atomic"
    "testing"
    "time"
)

type fakeTask struct {
    name     string
    interval time.Duration
    opts     TaskOpts
    runs     *atomic.Int64
    sleep    time.Duration
    retErr   error
}

func (f *fakeTask) Name() string                   { return f.name }
func (f *fakeTask) Interval() time.Duration        { return f.interval }
func (f *fakeTask) Opts() TaskOpts                 { return f.opts }
func (f *fakeTask) Run(ctx context.Context) error {
    if f.sleep > 0 {
        select {
        case <-time.After(f.sleep):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    f.runs.Add(1)
    return f.retErr
}

func newLogger() *slog.Logger {
    return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestBackGroundTaskRuntime_StartStop_RunsTask(t *testing.T) {
    runs := &atomic.Int64{}
    task := &fakeTask{
        name:     "tick",
        interval: 10 * time.Millisecond,
        opts:     TaskOpts{Overlap: OverlapOptEnqueue},
        runs:     runs,
        sleep:    1 * time.Millisecond,
    }

    rt := &BackGroundTaskRuntime{
        tasks:  []Task{task},
        logger: newLogger(),
    }

    ctx := context.Background()
    if err := rt.Start(ctx); err != nil {
        t.Fatalf("start: %v", err)
    }

    // Wait until at least 2 runs or timeout
    deadline := time.After(200 * time.Millisecond)
    ticker := time.NewTicker(5 * time.Millisecond)
    defer ticker.Stop()
    for {
        if runs.Load() >= 2 {
            break
        }
        select {
        case <-deadline:
            t.Fatalf("task did not run enough times, runs=%d", runs.Load())
        case <-ticker.C:
        }
    }

    stopCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()
    if err := rt.Stop(stopCtx); err != nil {
        t.Fatalf("stop: %v", err)
    }
}

func TestBackGroundTaskRuntime_NoTasks_NoOp(t *testing.T) {
    rt := &BackGroundTaskRuntime{logger: newLogger()}
    if err := rt.Start(context.Background()); err != nil {
        t.Fatalf("start: %v", err)
    }
    if err := rt.Stop(context.Background()); err != nil {
        t.Fatalf("stop: %v", err)
    }
}

func TestBackGroundTaskRuntime_NilReceiver_NoPanic(t *testing.T) {
    var rt *BackGroundTaskRuntime
    if err := rt.Start(context.Background()); err != nil {
        t.Fatalf("start on nil: %v", err)
    }
    if err := rt.Stop(context.Background()); err != nil {
        t.Fatalf("stop on nil: %v", err)
    }
}
