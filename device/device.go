package device

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
)

type Message interface {
	RoutingKey() string
	Data() interface{}
	Validate() error
}

type Instance struct {
	ID       string
	Name     string
	Language string
	Address  string
	Port     int
}

func (i *Instance) RoutingKey() string {
	return i.ID
}

func (i *Instance) Addr() string {
	return fmt.Sprintf("%s:%d", i.Address, i.Port)
}

type Device struct {
	ID     string
	Name   string
	Nets   []*labeled.Net
	Events []*labeled.Event
	*Instance
}

func (d *Device) EventMap() map[string]*petri.Transition {
	m := make(map[string]*petri.Transition)
	for _, net := range d.Nets {
		for e, t := range net.EventMap {
			m[e] = t.Transition
		}
	}
	return m
}

func New(id string, name string, nets []*labeled.Net) *Device {
	events := make([]*labeled.Event, 0)
	for _, net := range nets {
		events = append(events, net.Events...)
	}
	return &Device{
		ID:     id,
		Nets:   nets,
		Name:   name,
		Events: events,
	}
}

type ListItem struct {
	ID   string
	Name string
}

type Service interface {
	Load(ctx context.Context, devID string, handlers control.Handlers) (*Device, error)
	List(ctx context.Context) ([]*ListItem, error)
	Flush(ctx context.Context, dev *Device) (string, error)
}
