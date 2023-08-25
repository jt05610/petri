package labeled

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/marked"
	"log"
	"reflect"
	"strings"
)

type FieldType string

const (
	String  FieldType = "string"
	Number  FieldType = "number"
	Boolean FieldType = "boolean"
)

type Field struct {
	Name string
	Type FieldType
}

type Event struct {
	// The name of the event
	Name   string
	Fields []*Field
	// The data associated with the event
	Data map[string]interface{}
}

func firstCap(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// IsValid checks if the event valid aligns with what is expected by the fields.
func (e *Event) IsValid() bool {

	if e.Name == "" {
		return false
	}
	i := reflect.ValueOf(e.Data)
	// if it is a map[string]interface{} then check the fields directly
	if i.Kind() == reflect.Map {
		for _, f := range e.Fields {
			// get the map index of the field name
			mi := i.MapIndex(reflect.ValueOf(f.Name))
			// and the value of the map at that index
			mi = mi.Elem()
			// mi is an interface, so we need to check the type
			mi = reflect.ValueOf(mi.Interface())

			switch f.Type {
			case String:
				if mi.Kind() != reflect.String {
					return false
				}
			case Number:
				// check if the interface of the value of the map is a float64
				if mi.Kind() == reflect.Float64 || mi.Kind() == reflect.Float32 || mi.Kind() == reflect.Int || mi.Kind() == reflect.Int64 {
					return true
				}
				return false
			case Boolean:
				if mi.Kind() != reflect.Bool {
					return false
				}
			}
		}
		return true
	}

	// if it is a struct then check the fields directly
	if i.Kind() == reflect.Ptr {
		i = i.Elem()
	}
	if i.Kind() != reflect.Struct {
		return false
	}

	for _, f := range e.Fields {
		fbn := i.FieldByName(firstCap(f.Name))
		switch f.Type {
		case String:
			if fbn.Kind() != reflect.String {
				return false
			}
		case Number:
			if fbn.Kind() == reflect.Float64 || fbn.Kind() == reflect.Float32 || fbn.Kind() == reflect.Int || fbn.Kind() == reflect.Int64 {
				return true
			}
			return false
		case Boolean:
			if fbn.Kind() != reflect.Bool {
				return false
			}
		}
	}
	return true
}

type ColdTransition struct {
	*petri.Transition
	Handler
}

type Net struct {
	*marked.Net
	// Handlers are called when a transition is fired
	EventMap      map[string]*ColdTransition
	hot           map[string]bool
	notifications map[string][]*Notification
	Events        []*Event
	eventCh       chan *Event
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
		EventMap:      make(map[string]*ColdTransition),
		notifications: make(map[string][]*Notification),
		hot:           make(map[string]bool),
		eventCh:       make(chan *Event),
	}
	for _, t := range net.Transitions {
		n.hot[t.Name] = true
	}
	return n
}

func (n *Net) Channel() <-chan *Event {
	return n.eventCh
}

type Handler func(ctx context.Context, data *Event) (*Event, error)

func (n *Net) route(event string) (Handler, error) {
	if t, ok := n.EventMap[event]; ok {
		return t.Handler, nil
	}
	return nil, errors.New("no handler")
}

func sentenceCaseToSnakeCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
}

func (n *Net) AddEventHandler(event *Event, transition *petri.Transition, handler Handler) error {
	n.EventMap[sentenceCaseToSnakeCase(event.Name)] = &ColdTransition{
		Transition: transition,
		Handler:    handler,
	}
	if n.hot == nil {
		n.hot = make(map[string]bool)
	}
	n.hot[transition.Name] = false
	n.Events = append(n.Events, event)
	return nil
}

func (n *Net) AddHandler(event string, transition *petri.Transition, handler Handler) error {
	n.EventMap[event] = &ColdTransition{
		Transition: transition,
		Handler:    handler,
	}
	if n.hot == nil {
		n.hot = make(map[string]bool)
	}
	n.hot[transition.Name] = false
	n.Events = append(n.Events, &Event{
		Name: event,
	})

	return nil
}

type Getter func(ctx context.Context) (map[string]interface{}, error)

type Notification struct {
	Name string
	Getter
}

func (n *Net) AddNotification(name string, transition *petri.Transition, Getter func(ctx context.Context) (map[string]interface{}, error)) error {
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
	n.eventCh <- ev
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
					n.eventCh <- &Event{
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

func ValidSequence(net *Net, seq []*Event) bool {
	for _, e := range seq {
		if net.EventMap[e.Name].Transition == nil {
			return false
		}
	}
	testNet := &Net{
		Net:           net.Net.Copy(),
		EventMap:      make(map[string]*ColdTransition),
		notifications: make(map[string][]*Notification),
		hot:           make(map[string]bool),
		eventCh:       make(chan *Event),
	}
	for _, t := range testNet.Transitions {
		testNet.hot[t.Name] = true
	}
	for n, h := range net.EventMap {
		testNet.EventMap[n] = h
		testNet.hot[h.Transition.Name] = false
	}

	lnCh := testNet.Channel()
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case ev := <-lnCh:
				fmt.Printf("Received event %s\n", ev.Name)
			}
		}
	}()

	for _, event := range seq {
		err := testNet.Handle(context.Background(), event)
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	defer close(done)
	return true
}
