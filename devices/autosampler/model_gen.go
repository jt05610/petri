package autosampler

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	autosampler "github.com/jt05610/petri/devices/autosampler/proto"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/yaml"
	"log"
	"strconv"
	"sync/atomic"
)

func NewAutosampler(client autosampler.AutosamplerClient) *Autosampler {
	d := &Autosampler{
		client:      client,
		state:       new(atomic.Int32),
		stateChange: make(chan autosampler.InjectState),
	}
	d.state.Store(int32(autosampler.InjectState_Idle))
	return d
}

func (d *Autosampler) Load() *device.Device {
	srv := yaml.Service{}
	df, err := deviceYaml.Open("device.yaml")
	if err != nil {
		log.Fatal(err)
	}
	dev, err := srv.Load(df)
	if err != nil {
		log.Fatal(err)
	}
	ret, err := srv.ToNet(dev, d.Handlers())
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func (r *InjectRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "inject",
	}
}

func (r *InjectRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "inject" {
		return fmt.Errorf("expected event name inject, got %s", event.Name)
	}
	if event.Data["injectionVolume"] != nil {
		ds := event.Data["injectionVolume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.InjectionVolume = d
	}

	if event.Data["position"] != nil {
		ds := event.Data["position"].(string)

		r.Position = ds
	}

	if event.Data["airCushion"] != nil {
		ds := event.Data["airCushion"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.AirCushion = d
	}

	if event.Data["excessVolume"] != nil {
		ds := event.Data["excessVolume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.ExcessVolume = d
	}

	if event.Data["flushVolume"] != nil {
		ds := event.Data["flushVolume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.FlushVolume = d
	}

	if event.Data["needleDepth"] != nil {
		ds := event.Data["needleDepth"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.NeedleDepth = d
	}

	return nil
}

func (r *InjectResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "inject",
		Fields: []*labeled.Field{
			{
				Name: "injectionVolume",
				Type: "number",
			},

			{
				Name: "position",
				Type: "string",
			},

			{
				Name: "airCushion",
				Type: "number",
			},

			{
				Name: "excessVolume",
				Type: "number",
			},

			{
				Name: "flushVolume",
				Type: "number",
			},

			{
				Name: "needleDepth",
				Type: "number",
			},
		},
		Data: map[string]interface{}{
			"InjectionVolume": r.InjectionVolume,
			"Position":        r.Position,
			"AirCushion":      r.AirCushion,
			"ExcessVolume":    r.ExcessVolume,
			"FlushVolume":     r.FlushVolume,
			"NeedleDepth":     r.NeedleDepth,
		},
	}

	return ret
}

func (r *InjectResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "inject" {
		return fmt.Errorf("expected event name inject, got %s", event.Name)
	}
	return nil
}

func (r *InjectedRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "injected",
	}
}

func (r *InjectedRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "injected" {
		return fmt.Errorf("expected event name injected, got %s", event.Name)
	}
	return nil
}

func (r *InjectedResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "injected",
	}

	return ret
}

func (r *InjectedResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "injected" {
		return fmt.Errorf("expected event name injected, got %s", event.Name)
	}
	return nil
}

func (r *WaitForReadyRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "waitforready",
	}
}

func (r *WaitForReadyRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "waitforready" {
		return fmt.Errorf("expected event name waitforready, got %s", event.Name)
	}
	return nil
}

func (r *WaitForReadyResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "waitforready",
	}

	return ret
}

func (r *WaitForReadyResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "waitforready" {
		return fmt.Errorf("expected event name waitforready, got %s", event.Name)
	}
	return nil
}

func (d *Autosampler) Handlers() control.Handlers {
	return control.Handlers{

		"inject": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(InjectRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Inject(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"injected": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(InjectedRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Injected(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"waitforready": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(WaitForReadyRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.WaitForReady(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
