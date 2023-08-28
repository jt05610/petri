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

func NewTwoPositionThreeWayValve(txCh chan []byte, rxCh <-chan []byte) *TwoPositionThreeWayValve {
	d := &TwoPositionThreeWayValve{
		rxCh: rxCh,
		txCh: txCh,
	}
	return d
}

func (d *TwoPositionThreeWayValve) load() *device.Device {
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
	return nil
}

func (r *OpenAResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "open_a",
	}

	return ret
}

func (r *OpenAResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "open_a" {
		return fmt.Errorf("expected event name open_a, got %s", event.Name)
	}
	return nil
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
	return nil
}

func (r *OpenBResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "open_b",
	}

	return ret
}

func (r *OpenBResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "open_b" {
		return fmt.Errorf("expected event name open_b, got %s", event.Name)
	}
	return nil
}

func (d *TwoPositionThreeWayValve) Handlers() control.Handlers {
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
