package labeled

import (
	"context"
	"errors"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/marked"
	"sync"
)

type Event struct {
	// The name of the event
	Name string
	// The data associated with the event
	Data interface{}
}

type ColdTransition struct {
	*petri.Transition
	Handler
}
type Net struct {
	*marked.Net
	// Handlers are called when a transition is fired
	handlers      map[string]*ColdTransition
	hot           map[string]bool
	notifications map[string][]*Notification
	events        chan *Event
	mu            sync.Mutex
}

func (n *Net) Hot() []*petri.Transition {
	var hot []*petri.Transition
	for _, t := range n.Transitions {
		if n.hot[t.Name] {
			hot = append(hot, t)
		}
	}
	return hot
}

func New(net *marked.Net) *Net {
	n := &Net{
		Net:           net,
		handlers:      make(map[string]*ColdTransition),
		notifications: make(map[string][]*Notification),
		hot:           make(map[string]bool),
		events:        make(chan *Event),
	}
	for _, t := range net.Transitions {
		n.hot[t.Name] = true
	}
	return n
}

func (n *Net) Channel() <-chan *Event {
	return n.events
}

type Handler func(ctx context.Context, data *Event) (*Event, error)

func (n *Net) route(event string) (Handler, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if t, ok := n.handlers[event]; ok {
		err := n.Fire(t.Transition)
		if err != nil {
			return nil, err
		}
		return t.Handler, nil
	}
	return nil, errors.New("no handler")
}

func (n *Net) AddHandler(event string, transition *petri.Transition, handler Handler) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers[event] = &ColdTransition{
		Transition: transition,
		Handler:    handler,
	}
	n.hot[transition.Name] = false
	return nil
}

type Getter func(ctx context.Context) (interface{}, error)
type Notification struct {
	Name string
	Getter
}

func (n *Net) AddNotification(name string, transition *petri.Transition, Getter func(ctx context.Context) (interface{}, error)) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.notifications[transition.Name]; !ok {
		n.notifications[transition.Name] = make([]*Notification, 0)
	}
	n.notifications[transition.Name] = append(n.notifications[transition.Name], &Notification{
		Name:   name,
		Getter: Getter,
	})
	return nil
}

func (n *Net) Handle(ctx context.Context, event *Event) error {
	handler, err := n.route(event.Name)
	if err != nil {
		return err
	}
	ev, err := handler(ctx, event)
	n.events <- ev
	av := n.Available()
	nCold := 0
	for _, hot := range n.hot {
		if !hot {
			nCold++
		}
	}
	for len(av) > nCold {
		for _, t := range av {
			if !n.hot[t.Name] {
				continue
			}
			err := n.Fire(t)
			if nn, ok := n.notifications[t.Name]; ok {
				for _, h := range nn {
					d, err := h.Getter(ctx)
					if err != nil {
						return err
					}
					n.events <- &Event{
						Name: h.Name,
						Data: d,
					}
				}
			}
			if err != nil {
				return err
			}
		}
		av = n.Available()
	}
	if err != nil {
		return err
	}
	return nil
}
