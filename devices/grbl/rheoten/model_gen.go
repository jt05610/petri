package rheoten

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/yaml"
	"log"
	"strconv"
)

func NewTenPortRheodyneValve(client proto.GRBLServer) *TenPortRheodyneValve {
	d := &TenPortRheodyneValve{
		GRBLServer: client,
	}
	return d
}

func (d *TenPortRheodyneValve) load() *device.Device {
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

func (r *OpenARequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "open_a",
	}
}

func (r *OpenARequest) FromEvent(event *labeled.Event) error {
	if event.Name != "open_a" {
		return fmt.Errorf("expected event name open_a, got %s", event.Name)
	}
	if event.Data["delay"] != nil {
		ds := event.Data["delay"].(string)
		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Delay = d
	}

	return nil
}

func (r *OpenAResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "open_a",
		Fields: []*labeled.Field{
			{
				Name: "delay",
				Type: "number",
			},
		},
		Data: map[string]interface{}{
			"delay": r.Delay,
		},
	}

	return ret
}

func (r *OpenAResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "open_a" {
		return fmt.Errorf("expected event name open_a, got %s", event.Name)
	}
	return nil
}

type OpenBData struct {
	Delay float64 `json:"delay"`
}

func (r *OpenBRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "open_b",
	}
}

func (r *OpenBRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "open_b" {
		return fmt.Errorf("expected event name open_b, got %s", event.Name)
	}
	if event.Data["delay"] != nil {
		ds := event.Data["delay"].(string)
		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Delay = d
	}

	return nil
}

func (r *OpenBResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "open_b",
		Fields: []*labeled.Field{
			{
				Name: "delay",
				Type: "number",
			},
		},
		Data: map[string]interface{}{
			"delay": r.Delay,
		},
	}

	return ret
}

func (r *OpenBResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "open_b" {
		return fmt.Errorf("expected event name open_b, got %s", event.Name)
	}
	return nil
}

func (d *TenPortRheodyneValve) Handlers() control.Handlers {
	return control.Handlers{

		"open_a": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(OpenARequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.OpenA(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"open_b": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(OpenBRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.OpenB(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
