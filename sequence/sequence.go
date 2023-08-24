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

func (a *Action) ApplyParameters(params map[string]interface{}) {
	for _, p := range a.Parameters {
		if val, ok := params[p.Field.Name]; ok {
			p.Value = val
		}
	}
}

func (a *Action) ExtractParameters() {
	a.Parameters = make(map[string]*Parameter, len(a.Constants))
	for _, f := range a.Event.Fields {
		a.Parameters[f.Name] = &Parameter{
			Field: f,
		}
	}
	for _, c := range a.Constants {
		delete(a.Parameters, c.Name)
	}
}

type Step struct {
	*Action
}

type Sequence struct {
	ID          string
	Name        string
	Description string
	CurrentStep int
	Running     bool
	Steps       []*Step
}

func (s *Sequence) ApplyParameters(params map[string]interface{}) {
	for _, step := range s.Steps {
		step.ApplyParameters(params)
	}
}

func (s *Sequence) ExtractParameters() {
	for _, step := range s.Steps {
		step.ExtractParameters()
	}
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
