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

func NewOrganicPump() *OrganicPump {
	d := &OrganicPump{}
	return d
}

func (d *OrganicPump) load() *device.Device {
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

type InitializeData struct {
	Syringe_diameter float64 `json:"syringe_diameter"`
	Syringe_volume   float64 `json:"syringe_volume"`
	Steps_per_mm     float64 `json:"steps_per_mm"`
	Rate             float64 `json:"rate"`
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
	data, ok := event.Data.(*InitializeData)
	if !ok {
		return fmt.Errorf("expected data type InitializeData, got %T", event.Data)
	}

	r.Rate = data.Rate

	return nil
}

func (r *InitializeResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "initialize",
		Fields: []*labeled.Field{
			{
				Name: "syringe_diameter",
				Type: "number",
			},

			{
				Name: "syringe_volume",
				Type: "number",
			},

			{
				Name: "steps_per_mm",
				Type: "number",
			},

			{
				Name: "rate",
				Type: "number",
			},
		},
		Data: &InitializeData{
			Syringe_diameter: r.Syringe_diameter,
			Syringe_volume:   r.Syringe_volume,
			Steps_per_mm:     r.Steps_per_mm,
			Rate:             r.Rate,
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

type PumpData struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

func (r *PumpRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "pump",
	}
}

func (r *PumpRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "pump" {
		return fmt.Errorf("expected event name pump, got %s", event.Name)
	}
	data, ok := event.Data.(*PumpData)
	if !ok {
		return fmt.Errorf("expected data type PumpData, got %T", event.Data)
	}
	r.Volume = data.Volume

	r.Rate = data.Rate

	return nil
}

func (r *PumpResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "pump",
		Fields: []*labeled.Field{
			{
				Name: "volume",
				Type: "number",
			},

			{
				Name: "rate",
				Type: "number",
			},
		},
		Data: &PumpData{
			Volume: r.Volume,
			Rate:   r.Rate,
		},
	}

	return ret
}

func (r *PumpResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "pump" {
		return fmt.Errorf("expected event name pump, got %s", event.Name)
	}
	return nil
}

type PumpedData struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

func (r *PumpedRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "pumped",
	}
}

func (r *PumpedRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "pumped" {
		return fmt.Errorf("expected event name pumped, got %s", event.Name)
	}
	data, ok := event.Data.(*PumpedData)
	if !ok {
		return fmt.Errorf("expected data type PumpedData, got %T", event.Data)
	}
	r.Volume = data.Volume

	r.Rate = data.Rate

	return nil
}

func (r *PumpedResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "pumped",
		Fields: []*labeled.Field{
			{
				Name: "volume",
				Type: "number",
			},

			{
				Name: "rate",
				Type: "number",
			},
		},
		Data: &PumpedData{
			Volume: r.Volume,
			Rate:   r.Rate,
		},
	}

	return ret
}

func (r *PumpedResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "pumped" {
		return fmt.Errorf("expected event name pumped, got %s", event.Name)
	}
	return nil
}

func (d *OrganicPump) Handlers() control.Handlers {
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

		"pump": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(PumpRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Pump(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"pumped": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(PumpedRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Pumped(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
