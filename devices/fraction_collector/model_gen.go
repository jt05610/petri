package fracCollector

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	"github.com/jt05610/petri/labeled"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"github.com/jt05610/petri/yaml"
	"log"
	"strconv"
)

func NewFractionCollector(srv marlin.MarlinServer, layout *pipbot.Layout) *FractionCollector {
	d := &FractionCollector{
		MarlinServer: srv,
		Layout:       layout,
	}
	return d
}

func (d *FractionCollector) load() *device.Device {
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

func (r *CollectRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "collect",
	}
}

func (r *CollectRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "collect" {
		return fmt.Errorf("expected event name collect, got %s", event.Name)
	}
	if event.Data["wasteVol"] != nil {
		ds := event.Data["wasteVol"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.WasteVol = d
	}

	if event.Data["collectVol"] != nil {
		ds := event.Data["collectVol"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.CollectVol = d
	}

	return nil
}

func (r *CollectResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "collect",
		Fields: []*labeled.Field{
			{
				Name: "wastevol",
				Type: "number",
			},

			{
				Name: "collectvol",
				Type: "number",
			},
		},
		Data: map[string]interface{}{
			"WasteVol":   r.WasteVol,
			"CollectVol": r.CollectVol,
		},
	}

	return ret
}

func (r *CollectResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "collect" {
		return fmt.Errorf("expected event name collect, got %s", event.Name)
	}
	return nil
}

func (r *CollectedRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "collected",
	}
}

func (r *CollectedRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "collected" {
		return fmt.Errorf("expected event name collected, got %s", event.Name)
	}
	if event.Data["position"] != nil {
		ds := event.Data["position"].(string)

		r.Position = ds
	}

	if event.Data["grid"] != nil {
		ds := event.Data["grid"].(string)

		r.Grid = ds
	}

	return nil
}

func (r *CollectedResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "collected",
		Fields: []*labeled.Field{
			{
				Name: "position",
				Type: "string",
			},

			{
				Name: "grid",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"Position": r.Position,
			"Grid":     r.Grid,
		},
	}

	return ret
}

func (r *CollectedResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "collected" {
		return fmt.Errorf("expected event name collected, got %s", event.Name)
	}
	return nil
}

func (r *MoveToRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "moveto",
	}
}

func (r *MoveToRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "moveto" {
		return fmt.Errorf("expected event name moveto, got %s", event.Name)
	}
	if event.Data["position"] != nil {
		ds := event.Data["position"].(string)

		r.Position = ds
	}

	if event.Data["grid"] != nil {
		ds := event.Data["grid"].(string)

		r.Grid = ds
	}

	return nil
}

func (r *MoveToResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "moveto",
		Fields: []*labeled.Field{
			{
				Name: "position",
				Type: "string",
			},

			{
				Name: "grid",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"Position": r.Position,
			"Grid":     r.Grid,
		},
	}

	return ret
}

func (r *MoveToResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "moveto" {
		return fmt.Errorf("expected event name moveto, got %s", event.Name)
	}
	return nil
}

func (d *FractionCollector) Handlers() control.Handlers {
	return control.Handlers{

		"collect": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(CollectRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Collect(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"collected": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(CollectedRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Collected(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"moveto": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(MoveToRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.MoveTo(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
