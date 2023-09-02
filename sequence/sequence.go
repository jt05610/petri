package sequence

import (
	"context"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
	"strconv"
	"strings"
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
	for n, p := range a.Parameters {
		ret[n] = p.Value
	}
	for _, c := range a.Constants {
		ret[fLower(c.Name)] = c.Value
	}
	return ret
}

func (a *Action) ApplyParameters(params map[string]interface{}) error {
	for _, p := range a.Parameters {
		if val, ok := params[p.Field.ID]; ok {
			vm, ok := val.(map[string]interface{})
			if !ok {
				return labeled.ErrMissingParameter(p.Field, a.Event)
			}
			if v, ok := vm["value"]; ok {
				p.Value = v.(string)
				a.setParams++
			} else {
				return labeled.ErrMissingParameter(p.Field, a.Event)
			}

		} else {
			return labeled.ErrMissingParameter(p.Field, a.Event)
		}
	}
	a.Event.Data = a.ParameterMap()
	return nil
}

func fLower(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func (a *Action) ExtractParameters() {
	a.Parameters = make(map[string]*Parameter, len(a.Constants))
	for _, f := range a.Event.Fields {
		a.Parameters[fLower(f.Name)] = &Parameter{
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

func (s *Step) Command(to string) *control.Command {
	ev := s.Event
	return &control.Command{
		Event: ev,
		To:    to,
	}
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
	for _, dev := range s.Devices() {
		p, found := params[dev.ID]
		if !found {
			continue
		}
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		for i, step := range s.Steps {
			if stepParams, found := pm[strconv.Itoa(i)]; !found {
				continue
			} else {
				err := step.ApplyParameters(stepParams.(map[string]interface{}))
				if err != nil {
					return err
				}
			}
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
	Value string
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
