package events

type EventRouter struct {
	receiversByType map[string][]chan<- Event
}

func NewEventRouter() *EventRouter {
	return &EventRouter{receiversByType: make(map[string][]chan<- Event)}
}

func (r *EventRouter) Register(eventType string, receiver chan<- Event) {
	r.receiversByType[eventType] = append(r.receiversByType[eventType], receiver)
}

func (r *EventRouter) Route(event Event) {
	for _, receiver := range r.receiversByType[event.EventName()] {
		receiver <- event
	}
}
