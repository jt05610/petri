package camera

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/yaml"
	"log"
	"strconv"
)

func NewCamera() *Camera {
	d := &Camera{}
	return d
}

func (d *Camera) load() *device.Device {
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

func (r *CaptureRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "capture",
	}
}

func (r *CaptureRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "capture" {
		return fmt.Errorf("expected event name capture, got %s", event.Name)
	}
	if event.Data["duration"] != nil {
		ds := event.Data["duration"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Duration = d
	}

	if event.Data["interval"] != nil {
		ds := event.Data["interval"].(string)

		d, err := strconv.ParseFloat(ds, 64)
		if err != nil {
			return err
		}
		r.Interval = d
	}

	return nil
}

func (r *CaptureResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "capture",
		Fields: []*labeled.Field{
			{
				Name: "duration",
				Type: "number",
			},

			{
				Name: "interval",
				Type: "number",
			},
		},
		Data: map[string]interface{}{
			"Duration": r.Duration,
			"Interval": r.Interval,
		},
	}

	return ret
}

func (r *CaptureResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "capture" {
		return fmt.Errorf("expected event name capture, got %s", event.Name)
	}
	return nil
}

func (r *ImagesCapturedRequest) Event() *labeled.Event {
	return &labeled.Event{
		Name: "images_captured",
	}
}

func (r *ImagesCapturedRequest) FromEvent(event *labeled.Event) error {
	if event.Name != "images_captured" {
		return fmt.Errorf("expected event name images_captured, got %s", event.Name)
	}
	if event.Data["url"] != nil {
		ds := event.Data["url"].(string)

		r.Url = ds
	}

	return nil
}

func (r *ImagesCapturedResponse) Event() *labeled.Event {
	ret := &labeled.Event{
		Name: "images_captured",
		Fields: []*labeled.Field{
			{
				Name: "url",
				Type: "string",
			},
		},
		Data: map[string]interface{}{
			"url": r.Url,
		},
	}

	return ret
}

func (r *ImagesCapturedResponse) FromEvent(event *labeled.Event) error {
	if event.Name != "images_captured" {
		return fmt.Errorf("expected event name images_captured, got %s", event.Name)
	}
	return nil
}

func (d *Camera) Handlers() control.Handlers {
	return control.Handlers{

		"capture": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(CaptureRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.Capture(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},

		"images_captured": func(ctx context.Context, data *labeled.Event) (*labeled.Event, error) {
			req := new(ImagesCapturedRequest)
			err := req.FromEvent(data)
			if err != nil {
				return nil, err
			}
			resp, err := d.ImagesCaptured(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.Event(), nil
		},
	}
}
