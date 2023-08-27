package sequence

import (
	"context"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
)

type Constant struct {
	*labeled.Field
	Value interface{}
}

type Action struct {
	Constants  []*Constant
	Parameters map[string]*Parameter
	setParams  int
	*device.Device
	Event *labeled.Event
}

func (a *Action) ParameterMap() map[string]interface{} {
	ret := make(map[string]interface{})
	for _, p := range a.Parameters {
		ret[p.Field.Name] = p.Value
	}
	return ret
}

func (a *Action) ApplyParameters(params map[string]interface{}) error {
	for _, p := range a.Parameters {
		if val, ok := params[p.Field.Name]; ok {
			p.Value = val
		} else {
			return labeled.ErrMissingParameter(p.Field, a.Event)
		}
	}
	return nil
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
	ID             string
	InitialMarking control.Marking
	NetID          string
	Net            *labeled.Net
	Name           string
	Description    string
	CurrentStep    int
	Running        bool
	Steps          []*Step
}

func (s *Sequence) ApplyNet(net *marked.Net) error {
	s.Net = labeled.New(net)
	s.InitialMarking = net.MarkingMap()
	return nil
}

func (s *Sequence) ApplyParameters(params map[string]interface{}) error {
	for _, step := range s.Steps {
		err := step.ApplyParameters(params)
		if err != nil {
			return err
		}
	}
	return nil
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

func (s *Sequence) Events() []*labeled.Event {
	ret := make([]*labeled.Event, len(s.Steps))
	for _, step := range s.Steps {
		ret = append(ret, step.Event)
	}
	return ret
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
