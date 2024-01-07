package device

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
)

// Message is a message sent to or from a device. Incoming messages are instructions to the device sent by the
// broker, and outgoing messages are events sent by the device to the broker along with any data collected by the
// device while handling the event.
type Message interface {
	// RoutingKey is the routing key used to send the message to the device.
	RoutingKey() string
	// Data is the data sent to or from the device.
	Data() interface{}
	// Validate validates the message, and raises an error if the message is invalid.
	Validate() error
}

// Instance is an implementation of a Device. Instances are communicated with by the broker via AMQP. Instances can also
// be communicated with via GraphQL.
type Instance struct {
	// deviceID is the ID of the device which the Instance implements.
	deviceID string
	// ID is the unique identifier for the Instance.
	ID string
	// Name is the name of the Instance.
	Name string
	// Language is the language which the Instance is implemented in.
	Language string
	// Address is the IP address of the Instance.
	Address string
	// Port is the port which the Instance is listening on.
	Port int
	// Device is the device which the Instance implements
	Device *Device
}

// RoutingKey returns the routing key used to send messages to the Instance.
func (i *Instance) RoutingKey() string {
	return i.ID
}

// Addr returns the address of the Instance.
func (i *Instance) Addr() string {
	return fmt.Sprintf("%s:%d", i.Address, i.Port)
}

// Device is a device which implements code handlers for Events associated with one or many Nets. A Device
// is a collection of Nets, and a collection of Events, and a collection of Instances.
type Device struct {
	// ID is the unique identifier for the device.
	ID string
	// Name is the name of the device.
	Name string
	// Nets is the collection of labeled Petri Nets which the device implements.
	Nets []*labeled.Net
}

// EventMap returns a map of event names to the Transitions which should be fired when the event is received.
func (d *Device) EventMap() map[string]*petri.Transition {
	m := make(map[string]*petri.Transition)
	for _, net := range d.Nets {
		for e, t := range net.EventMap {
			m[e] = t.Transition
		}
	}
	return m
}

// Events is the collection of Events which the device should handle.
func (d *Device) Events() []*labeled.Event {
	var events []*labeled.Event
	for _, net := range d.Nets {
		events = append(events, net.Events...)
	}
	return events
}

// New returns a new Device with the given ID, name, and Nets.
func New(id string, name string, nets []*labeled.Net) *Device {
	return &Device{
		ID:   id,
		Nets: nets,
		Name: name,
	}
}

// ListItem is a list item for a Device.
type ListItem struct {
	// ID is the unique identifier for the Device.
	ID string
	// Name is the name of the Device.
	Name string
}

// Service is a service for managing Devices.
type Service interface {
	// Load loads a Device with the given ID, and returns the Device.
	Load(ctx context.Context, devID string, handlers control.Handlers) (*Device, error)
	List(ctx context.Context) ([]*ListItem, error)
	Flush(ctx context.Context, dev *Device, inst *Instance) (string, error)
}
