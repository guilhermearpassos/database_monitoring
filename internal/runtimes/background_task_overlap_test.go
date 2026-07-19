package runtimes

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type overlapFakeTask struct {
	name     string
	interval time.Duration
	opts     TaskOpts
	// config
	runDur time.Duration // how long Run should take when not canceled
	// optional test controls (nil unless a test sets them)
	blockUntil <-chan struct{} // if set, Run will wait for this to be closed before finishing (unless ctx canceled)
	startedCh  chan time.Time  // if set, Run will non-blockingly notify start times here

	// observability
	running       atomic.Int64 // current concurrent runs
	maxConcurrent atomic.Int64
	starts        atomic.Int64
	cancels       atomic.Int64

	mu         sync.Mutex
	startTimes []time.Time // timestamps of Run start, to analyze spacing
}

func (f *overlapFakeTask) Name() string            { return f.name }
func (f *overlapFakeTask) Interval() time.Duration { return f.interval }
func (f *overlapFakeTask) Opts() TaskOpts          { return f.opts }
func (f *overlapFakeTask) Run(ctx context.Context) error {
	cur := f.running.Add(1)
	// track max concurrent
	for {
		max := f.maxConcurrent.Load()
		if cur <= max || f.maxConcurrent.CompareAndSwap(max, cur) {
			break
		}
	}
	f.starts.Add(1)

	now := time.Now()
	f.mu.Lock()
	f.startTimes = append(f.startTimes, now)
	f.mu.Unlock()
	// non-blocking notify
	if f.startedCh != nil {
		select {
		case f.startedCh <- now:
		default:
		}
	}

	// Simulate work or wait for cancellation
	var timer *time.Timer
	if f.runDur > 0 {
		timer = time.NewTimer(f.runDur)
		defer timer.Stop()
	}
	var canceled bool
	if f.blockUntil != nil {
		select {
		case <-ctx.Done():
			canceled = true
		case <-f.blockUntil:
		}
	} else if timer != nil {
		select {
		case <-ctx.Done():
			canceled = true
		case <-timer.C:
		}
	} else {
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}
	}

	if canceled {
		f.cancels.Add(1)
	}

	f.running.Add(-1)
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newRuntimeWithTask(t Task) *BackGroundTaskRuntime {
	return &BackGroundTaskRuntime{
		tasks:  []Task{t},
		logger: discardLogger(),
	}
}

func waitFor(predicate func() bool, timeout time.Duration) bool {
	deadline := time.NewTimer(timeout)
	ticker := time.NewTicker(5 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		if predicate() {
			return true
		}
		select {
		case <-deadline.C:
			return false
		case <-ticker.C:
		}
	}
}

func TestOverlap_Allow_AllowsConcurrency(t *testing.T) {
	task := &overlapFakeTask{
		name:     "allow",
		interval: 10 * time.Millisecond,
		opts:     TaskOpts{Overlap: OverlapOptAllow},
		runDur:   60 * time.Millisecond,
	}
	rt := newRuntimeWithTask(task)
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Wait until we observe at least 2 concurrent runs
	ok := waitFor(func() bool { return task.maxConcurrent.Load() >= 2 }, 300*time.Millisecond)
	stopCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := rt.Stop(stopCtx); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if !ok {
		t.Fatalf("expected concurrent runs with allow; maxConcurrent=%d", task.maxConcurrent.Load())
	}
}

func TestOverlap_Drop_NeverOverlaps(t *testing.T) {
	task := &overlapFakeTask{
		name:     "drop",
		interval: 15 * time.Millisecond,
		opts:     TaskOpts{Overlap: OverlapOptDrop},
		runDur:   60 * time.Millisecond, // much longer than interval, so many ticks happen while running
	}
	rt := newRuntimeWithTask(task)
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Let it run for a short window
	time.Sleep(220 * time.Millisecond)

	stopCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := rt.Stop(stopCtx); err != nil {
		t.Fatalf("stop: %v", err)
	}

	if task.maxConcurrent.Load() != 1 {
		t.Fatalf("drop mode should not overlap; maxConcurrent=%d", task.maxConcurrent.Load())
	}

	// Upper bound on runs: roughly window/runDur + 2 (start + maybe one more tick after finish)
	// This ensures many ticks were dropped while a run was in progress.
	if n := task.starts.Load(); n > 6 { // very conservative upper bound for this short window
		t.Fatalf("expected few runs with drop mode; got %d", n)
	}
}

func TestOverlap_Enqueue_QueuesOnePending_NoOverlap(t *testing.T) {
	interval := 20 * time.Millisecond
	runDur := 80 * time.Millisecond // slightly longer to build clearer backlog
	task := &overlapFakeTask{
		name:     "enqueue",
		interval: interval,
		opts:     TaskOpts{Overlap: OverlapOptEnqueue},
		runDur:   runDur,
	}
	rt := newRuntimeWithTask(task)
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Let it run for long enough to build backlog
	time.Sleep(280 * time.Millisecond)

	stopCtx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	if err := rt.Stop(stopCtx); err != nil {
		t.Fatalf("stop: %v", err)
	}

	// No overlaps expected
	if task.maxConcurrent.Load() != 1 {
		t.Fatalf("enqueue mode should not overlap; maxConcurrent=%d", task.maxConcurrent.Load())
	}

	// Analyze inter-start deltas: with enqueue, after a long run, the next should start almost immediately
	task.mu.Lock()
	times := append([]time.Time(nil), task.startTimes...)
	task.mu.Unlock()

	if len(times) < 3 {
		t.Logf("insufficient samples for delta analysis; starts=%d", len(times))
		return
	}

	var minDelta time.Duration = time.Hour
	for i := 1; i < len(times); i++ {
		d := times[i].Sub(times[i-1])
		if d < minDelta {
			minDelta = d
		}
	}
	// Allow for OS timer and scheduler jitter (especially on Windows). Expect at least one near-immediate start
	// where queued tick fires right after a long run completes. Consider anything within max(60ms, 2*interval) as immediate.
	jitter := 60 * time.Millisecond
	threshold := 2 * interval
	if jitter > threshold {
		threshold = jitter
	}
	if minDelta > threshold {
		t.Fatalf("expected an immediate-ish start due to queued tick; min inter-start delta=%v (threshold=%v, interval=%v)", minDelta, threshold, interval)
	}
}

func TestOverlap_CancelCurrent_CancelsAndRestarts(t *testing.T) {
	// Task will run for a long time unless its context is canceled
	task := &overlapFakeTask{
		name:     "cancel",
		interval: 20 * time.Millisecond,
		opts:     TaskOpts{Overlap: OverlapOptCancelCurrent},
		runDur:   500 * time.Millisecond,
	}
	rt := newRuntimeWithTask(task)
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Give it some time to accumulate cancels due to rapid ticks
	time.Sleep(200 * time.Millisecond)

	stopCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := rt.Stop(stopCtx); err != nil {
		t.Fatalf("stop: %v", err)
	}

	if task.maxConcurrent.Load() != 1 {
		t.Fatalf("cancel-current should not overlap; maxConcurrent=%d", task.maxConcurrent.Load())
	}
	if task.cancels.Load() == 0 {
		t.Fatalf("expected at least one cancellation in cancel-current mode; cancels=%d", task.cancels.Load())
	}
}
