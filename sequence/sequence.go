package sequence

import (
	"context"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
)

type Constant struct {
	*labeled.Field
	Value interface{}
}

type Action struct {
	Constants  []*Constant
	Parameters map[string]*Parameter
	*device.Device
	Event *labeled.Event
}

func NewAction(constants []*Constant, d *device.Device, e *labeled.Event) *Action {
	a := &Action{
		Constants: constants,
		Device:    d,
		Event:     e,
	}
	a.Parameters = a.parameters()
	return a
}

func (a *Action) parameters() map[string]*Parameter {
	ret := make(map[string]*Parameter, len(a.Constants))
	for _, f := range a.Event.Fields {
		ret[f.Name] = &Parameter{
			Field: f,
		}
	}
	for _, c := range a.Constants {
		delete(ret, c.Name)
	}
	return ret
}

type Step struct {
	*Action
}

type Sequence struct {
	ID          string
	Name        string
	Description string
	Steps       []*Step
	Parameters  []*Parameter
}

func (s *Sequence) SetInstance(deviceID string, i *device.Instance) {
	for _, step := range s.Steps {
		if step.Device.ID == deviceID {
			step.Device.Instance = i
		}
	}
}

type Parameter struct {
	Field *labeled.Field
	Value interface{}
}

type Config struct {
	DeviceInstance map[string]*device.Instance
	Steps          []*Step
}

func New(cfg *Config) *Sequence {
	s := &Sequence{
		Steps: cfg.Steps,
	}
	for _, d := range s.Devices() {
		if inst, ok := cfg.DeviceInstance[d.ID]; !ok {
			panic("missing device instance")
		} else {
			d.Instance = inst
		}
	}
	return s
}

func (s *Sequence) Devices() []*device.Device {
	seen := make(map[string]*device.Device)
	ret := make([]*device.Device, 0)
	for _, step := range s.Steps {
		if _, ok := seen[step.Device.ID]; !ok {
			seen[step.Device.ID] = step.Device
			ret = append(ret, step.Device)
		}
	}
	return ret
}

type ListItem struct {
	ID          string
	Name        string
	Description string
}

type Service interface {
	// Load loads the run from the database
	Load(ctx context.Context, id string) (*Sequence, error)
	List(ctx context.Context) ([]*ListItem, error)
}
