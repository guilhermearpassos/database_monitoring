package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	chSizeCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "sqlsights",
		Subsystem:   "",
		Name:        "router_channel_size",
		Help:        "",
		ConstLabels: nil,
	}, []string{"eventType", "channelName", "target"})
	chDropCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "sqlsights",
		Subsystem:   "",
		Name:        "router_channel_dropped_total",
		Help:        "number of events dropped because receiver channel was full",
		ConstLabels: nil,
	}, []string{"eventType", "channelName", "target"})
	register = sync.OnceFunc(func() {
		prometheus.MustRegister(chSizeCounter)
		prometheus.MustRegister(chDropCounter)
	})
)

type EventRouter struct {
	receiversByType map[string][]chan<- Event
	chNames         map[string][]string
	target          string
}

func NewEventRouter(target string) *EventRouter {
	return &EventRouter{
		receiversByType: make(map[string][]chan<- Event),
		chNames:         make(map[string][]string),
		target:          target,
	}
}

func (r *EventRouter) Register(eventType string, receiver chan<- Event, channelName string) {
	r.receiversByType[eventType] = append(r.receiversByType[eventType], receiver)
	r.chNames[eventType] = append(r.chNames[eventType], channelName)
}

func (r *EventRouter) Route(event Event) {
	evType := event.EventName()
	receivers := r.receiversByType[evType]
	chNames := r.chNames[evType]
	for i := 0; i < len(receivers); i++ {
		ch := receivers[i]
		name := chNames[i]
		select {
		case ch <- event:
			// delivered
		default:
			// drop and record
			chDropCounter.WithLabelValues(evType, name, r.target).Inc()
			// best-effort warn; no ctx available on router, so log without context
			telemetry.Info(context.Background(), "router dropped event due to full channel", "event_type", evType, "channel", name, "target", r.target)
		}
	}
}
func (r *EventRouter) StartMetrics(ctx context.Context) {
	register()
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	defer func() {
		if rec := recover(); rec != nil {
			telemetry.Error(context.Background(), fmt.Errorf("%v", rec), "router metrics panic recovered")
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for evType, channels := range r.receiversByType {
				chNames := r.chNames[evType]
				for i := 0; i < len(channels); i++ {
					ch := channels[i]
					name := chNames[i]
					chSizeCounter.WithLabelValues(evType, name, r.target).Set(float64(len(ch)))
				}
			}
		}
	}
}
