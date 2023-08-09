package control

import (
	"context"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
)

type Device struct {
	ID string
	*marked.Net
	handlers Handlers
	mu       chan struct{}
}

type DeviceMap map[string]*Device

type Controller struct {
	devices DeviceMap
}

type Handlers map[string]labeled.Handler

func (h Handlers) Handle(ctx context.Context, data *Command) (*Event, error) {
	res, err := (h)[data.Event.Name](ctx, data.Event)
	if err != nil {
		return nil, err
	}
	return &Event{Event: res}, nil
}
