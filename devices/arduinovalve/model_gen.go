package main

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/yaml"
	"log"
)

func NewMixingValve() *MixingValve {
	d := &MixingValve{}
	return d
}

func (d *MixingValve) load() *device.Device {
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

func (r *InitializeRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "initialize",
	}
}

func (r *InitializeRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "initialize" {
		return fmt.Errorf("expected event name initialize, got %s", event.Name)
	}
	if event.Data["components"] != nil {
		ds := event.Data["components"].(string)

		r.Components = ds
	}

	return nil
}

func (r *InitializeResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "initialize",
		Fields: []*labeled.Field{
			{
				Name: "components",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"Components": r.Components,
		},
	}

	return ret
}

func (r *InitializeResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "initialize" {
		return fmt.Errorf("expected event name initialize, got %s", event.Name)
	}
	return nil
}

func (r *MixRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "mix",
	}
}

func (r *MixRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "mix" {
		return fmt.Errorf("expected event name mix, got %s", event.Name)
	}
	if event.Data["proportions"] != nil {
		ds := event.Data["proportions"].(string)

		r.Proportions = ds
	}

	return nil
}

func (r *MixResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "mix",
		Fields: []*labeled.Field{
			{
				Name: "proportions",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"Proportions": r.Proportions,
		},
	}

	return ret
}

func (r *MixResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "mix" {
		return fmt.Errorf("expected event name mix, got %s", event.Name)
	}
	return nil
}

func (r *MixedRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "mixed",
	}
}

func (r *MixedRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "mixed" {
		return fmt.Errorf("expected event name mixed, got %s", event.Name)
	}
	if event.Data["proportions"] != nil {
		ds := event.Data["proportions"].(string)

		r.Proportions = ds
	}

	return nil
}

func (r *MixedResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "mixed",
		Fields: []*labeled.Field{
			{
				Name: "proportions",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"Proportions": r.Proportions,
		},
	}

	return ret
}

func (r *MixedResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "mixed" {
		return fmt.Errorf("expected event name mixed, got %s", event.Name)
	}
	return nil
}

func (d *MixingValve) Handlers() control.Handlers {
	return control.Handlers{

		"initialize": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(InitializeRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Initialize(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"mix": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(MixRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Mix(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"mixed": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(MixedRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Mixed(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
