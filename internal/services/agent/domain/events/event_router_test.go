package events

import (
	"context"
	"testing"
)

type FakeEvent struct {
}

func (f FakeEvent) EventName() string {
	return "fakeEvent"
}

func (f FakeEvent) Context() context.Context {
	return context.Background()
}

type FakeEvent2 struct {
}

func (f FakeEvent2) EventName() string {
	return "fakeEvent2"
}
func (f FakeEvent2) Context() context.Context {
	return context.Background()
}
func TestEventRouter_Route(t *testing.T) {
	tests := []struct {
		name            string
		receiversByType map[string][]chan Event
		chNames         map[string][]string
		events          []Event
		expected        int
	}{
		{
			name: "single event single receiver",
			receiversByType: map[string][]chan Event{
				"fakeEvent": {make(chan Event, 1)},
			},
			chNames: map[string][]string{
				"fakeEvent": {"receiver1"},
			},
			events:   []Event{FakeEvent{}},
			expected: 1,
		},
		{
			name: "single event multiple receivers",
			receiversByType: map[string][]chan Event{
				"fakeEvent": {make(chan Event, 1), make(chan Event, 1)},
			},
			chNames: map[string][]string{
				"fakeEvent": {"receiver1", "receiver2"},
			},
			events:   []Event{FakeEvent{}},
			expected: 2,
		},
		{
			name: "multiple events multiple receivers",
			receiversByType: map[string][]chan Event{
				"fakeEvent":  {make(chan Event, 2), make(chan Event, 2)},
				"fakeEvent2": {make(chan Event, 2)},
			},
			chNames: map[string][]string{
				"fakeEvent":  {"receiver1", "receiver2"},
				"fakeEvent2": {"receiver3"},
			},
			events:   []Event{FakeEvent{}, FakeEvent2{}},
			expected: 3,
		},
		{
			name: "no matching receivers",
			receiversByType: map[string][]chan Event{
				"fakeEvent2": {make(chan Event, 1)},
			},
			chNames: map[string][]string{
				"fakeEvent2": {"receiver1"},
			},
			events:   []Event{FakeEvent{}},
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewEventRouter("")
			for eventType, chs := range tt.receiversByType {
				for i := 0; i < len(chs); i++ {
					name := tt.chNames[eventType][i]
					r.Register(eventType, chs[i], name)
				}

			}
			go func() {
				for _, ev := range tt.events {
					r.Route(ev)
				}
			}()
			got := 0
			for got < tt.expected {
				for _, chs := range tt.receiversByType {
					for _, ch := range chs {
						select {
						case <-ch:
							got++
						default:
							continue
						}
					}
				}
			}
		})
	}
}
