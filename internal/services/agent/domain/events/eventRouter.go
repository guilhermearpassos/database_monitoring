package events

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type EventRouter struct {
	receiversByType map[string][]chan<- Event
	chNames         map[string][]string
	chSizeCounter   *prometheus.GaugeVec
}

func NewEventRouter() *EventRouter {
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "sqlsights",
		Subsystem:   "",
		Name:        "router_channel_size",
		Help:        "",
		ConstLabels: nil,
	}, []string{"eventType", "channelName"})
	prometheus.MustRegister(vec)
	return &EventRouter{
		receiversByType: make(map[string][]chan<- Event),
		chNames:         make(map[string][]string),
		chSizeCounter:   vec,
	}
}

func (r *EventRouter) Register(eventType string, receiver chan<- Event, channelName string) {
	r.receiversByType[eventType] = append(r.receiversByType[eventType], receiver)
	r.chNames[eventType] = append(r.chNames[eventType], channelName)
}

func (r *EventRouter) Route(event Event) {
	for _, receiver := range r.receiversByType[event.EventName()] {
		receiver <- event
	}
}
func (r *EventRouter) StartMetrics(ctx context.Context) {
	t := time.NewTicker(10 * time.Second)
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
					r.chSizeCounter.WithLabelValues(evType, name).Set(float64(len(ch)))
				}
			}
		}
	}
}
