package main

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

func NewOrganicPump(client proto.GRBLClient) *OrganicPump {
	d := &OrganicPump{
		GRBLClient: client,
	}
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

func (r *InitializeRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "initialize",
	}
}

func (r *InitializeRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "initialize" {
		return fmt.Errorf("expected event name initialize, got %s", event.Name)
	}
	if event.Data["syringe_diameter"] != nil {
		ds := event.Data["syringe_diameter"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.SyringeDiameter = d
	}

	if event.Data["syringe_volume"] != nil {
		ds := event.Data["syringe_volume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.SyringeVolume = d
	}

	if event.Data["steps_per_mm"] != nil {
		ds := event.Data["steps_per_mm"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.StepsPerMM = d
	}

	if event.Data["rate"] != nil {
		ds := event.Data["rate"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Rate = d
	}

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
		Data: map[string]interface{}{
			"syringe_diameter": r.SyringeDiameter,
			"syringe_volume":   r.SyringeVolume,
			"steps_per_mm":     r.StepsPerMM,
			"rate":             r.Rate,
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

func (r *PumpRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "pump",
	}
}

func (r *PumpRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "pump" {
		return fmt.Errorf("expected event name pump, got %s", event.Name)
	}
	if event.Data["volume"] != nil {
		ds := event.Data["volume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Volume = d
	}

	if event.Data["rate"] != nil {
		ds := event.Data["rate"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Rate = d
	}

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
		Data: map[string]interface{}{
			"Volume": r.Volume,
			"Rate":   r.Rate,
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

func (r *PumpedRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "pumped",
	}
}

func (r *PumpedRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "pumped" {
		return fmt.Errorf("expected event name pumped, got %s", event.Name)
	}
	if event.Data["volume"] != nil {
		ds := event.Data["volume"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Volume = d
	}

	if event.Data["rate"] != nil {
		ds := event.Data["rate"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Rate = d
	}

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
		Data: map[string]interface{}{
			"Volume": r.Volume,
			"Rate":   r.Rate,
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
