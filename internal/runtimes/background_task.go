package runtimes

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"sync"
	"time"
)

type OverlapOpt string

const (
	OverlapOptEnqueue       OverlapOpt = "enqueue"
	OverlapOptDrop          OverlapOpt = "drop"
	OverlapOptCancelCurrent OverlapOpt = "cancel-current"
	OverlapOptAllow         OverlapOpt = "allow"
)

var (
	backgroundHistogramVec *prometheus.HistogramVec
	backgroundCounterVec   *prometheus.CounterVec
)

func init() {
	backgroundHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   "sqlsights",
			Subsystem:   "",
			Name:        "background_job_seconds",
			Help:        "",
			ConstLabels: nil,
			Buckets: []float64{
				0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20, 30, 60, 120, 180,
			},
		},
		[]string{"task_name", "success"},
	)
	backgroundCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "sqlsights",
			Subsystem: "",
			Name:      "background_job_count",
			Help:      "",
		},
		[]string{"task_name", "success"},
	)
	prometheus.MustRegister(backgroundHistogramVec, backgroundCounterVec)
}

type TaskOpts struct {
	Overlap OverlapOpt
}

type Task interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
	Opts() TaskOpts
}
type BackGroundTaskRuntime struct {
	tasks   []Task
	wg      sync.WaitGroup
	started bool
	cancel  context.CancelFunc
	logger  *slog.Logger
}

var _ Runtime = (*BackGroundTaskRuntime)(nil)

func (b *BackGroundTaskRuntime) Start(ctx context.Context) error {
	if b == nil {
		return nil
	}
	if len(b.tasks) == 0 {
		return nil
	}
	if b.started {
		return nil
	}
	b.started = true
	runCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	for _, task := range b.tasks {
		if task == nil {
			continue
		}
		b.wg.Add(1)
		go func(task Task) {
			defer b.wg.Done()
			b.runTaskLoop(runCtx, task)
		}(task)
	}
	return nil
}
func (b *BackGroundTaskRuntime) Stop(ctx context.Context) error {
	if b == nil {
		return nil
	}
	if b.cancel == nil {
		return nil
	}
	b.cancel()
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		b.started = false
		b.cancel = nil
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		b.logger.Warn("context canceled %v", ctx.Err())
		return nil
	}
}

func (b *BackGroundTaskRuntime) Type() RuntimeType {
	return BackGroundTask
}

func (b *BackGroundTaskRuntime) runTaskLoop(ctx context.Context, task Task) {
	interval := task.Interval()
	if interval <= 0 {
		panic(fmt.Errorf("background task interval cannot be negative - %s: %v", task.Name(), interval))
	}
	opts := task.Opts()
	switch opts.Overlap {
	case OverlapOptEnqueue:
		b.runEnqueueLoop(ctx, task)
	case OverlapOptDrop:
		b.runDropOnOverlapLoop(ctx, task)
	case OverlapOptCancelCurrent:
		b.runCancelOnOverlapLoop(ctx, task)
	case OverlapOptAllow:
		b.runAllowOverlapLoop(ctx, task)
	default:
		panic(fmt.Errorf("invalid overlap option on task %s: %s", task.Name(), opts.Overlap))
	}

}
func (b *BackGroundTaskRuntime) runTask(ctx context.Context, task Task) {
	start := time.Now()
	success := false

	defer func() {
		elapsed := time.Since(start)

		if r := recover(); r != nil {
			// log panic here
			success = false
		}

		successLabel := fmt.Sprintf("%t", success)
		backgroundHistogramVec.WithLabelValues(task.Name(), successLabel).Observe(elapsed.Seconds())
		backgroundCounterVec.WithLabelValues(task.Name(), successLabel).Inc()
	}()

	err := task.Run(ctx)
	success = err == nil
	if err != nil {
		b.logger.Error(fmt.Sprintf("running task %s: %s", task.Name(), err.Error()))
	}
	return
}

func (b *BackGroundTaskRuntime) runEnqueueLoop(ctx context.Context, task Task) {

	queue := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(task.Interval())
		defer ticker.Stop()
		queue <- struct{}{}
		for {
			select {
			case <-ticker.C:
				select {
				case queue <- struct{}{}:
				default:
				}

			case <-ctx.Done():
				return

			}
		}
	}()
	for {
		select {
		case <-queue:
			b.runTask(ctx, task)
		case <-ctx.Done():
			return
		}
	}
}
func (b *BackGroundTaskRuntime) runDropOnOverlapLoop(ctx context.Context, task Task) {
	ticker := time.NewTicker(task.Interval())
	defer ticker.Stop()
	defer ticker.Stop()
	running := false
	var mu sync.Mutex
	start := func() {
		mu.Lock()
		if running {
			//drop this run
			mu.Unlock()
			return
		}
		running = true
		mu.Unlock()
		go func() {
			defer func() {
				mu.Lock()
				running = false
				mu.Unlock()
			}()
			b.runTask(ctx, task)
		}()
	}
	start()
	for {
		select {
		case <-ticker.C:
			start()
		case <-ctx.Done():
			return
		}
	}

}
func (b *BackGroundTaskRuntime) runCancelOnOverlapLoop(ctx context.Context, task Task) {
	ticker := time.NewTicker(task.Interval())
	defer ticker.Stop()
	running := false
	var mu sync.Mutex
	var cancel context.CancelFunc
	start := func() {
		mu.Lock()
		if running {
			if cancel != nil {
				cancel()
				cancel = nil
			}
			mu.Unlock()
		}
		var runCtx context.Context
		runCtx, cancel = context.WithCancel(ctx)
		running = true
		mu.Unlock()
		done := make(chan struct{}, 1)
		go func() {
			b.runTaskLoop(runCtx, task)
			done <- struct{}{}
		}()
		select {
		case <-done:
			return
		case <-done:
			mu.Lock()
			if running {
				running = false
			}
			if cancel != nil {
				cancel()
				cancel = nil
			}
			mu.Unlock()
		}
	}

	start()
	select {
	case <-ticker.C:
		start()
	case <-ctx.Done():
		return
	}
}
func (b *BackGroundTaskRuntime) runAllowOverlapLoop(ctx context.Context, task Task) {
	ticker := time.NewTicker(task.Interval())
	defer ticker.Stop()
	start := func() {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			b.runTask(ctx, task)
		}()
	}
	start()
	for {
		select {
		case <-ticker.C:
			start()
		case <-ctx.Done():
			return
		}
	}
}
