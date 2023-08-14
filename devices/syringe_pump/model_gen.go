package main

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/proto/v1"
	"github.com/jt05610/petri/yaml"
	"log"
)

func NewSyringePump(client modbus.ModbusClient) *SyringePump {
	d := &SyringePump{client: client}
	return d
}

func (d *SyringePump) load() *device.Device {
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
	return nil
}

func (r *InitializeResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "initialize",
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

func (r *StopRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "stop",
	}
}

func (r *StopRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "stop" {
		return fmt.Errorf("expected event name stop, got %s", event.Name)
	}
	return nil
}

func (r *StopResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "stop",
	}

	return ret
}

func (r *StopResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "stop" {
		return fmt.Errorf("expected event name stop, got %s", event.Name)
	}
	return nil
}

func (d *SyringePump) Handlers() control.Handlers {
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

		"stop": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(StopRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Stop(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
