package device

import (
	"context"
	"errors"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp"
	"github.com/rabbitmq/amqp091-go"
	"io"
)

type Device struct {
	Net *petri.Net
	*amqp.Server
	Remotes map[string]string
}

type Service interface {
	Load(ctx context.Context, r io.Reader) (*Device, error)
	Save(ctx context.Context, w io.Writer, d *Device) error
}

func (d *Device) WithHandlers(handlers amqp.Handlers) *Device {
	d.RegisterHandlers(handlers)
	return d
}

func (d *Device) Connect(url string) error {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return err
	}
	if d.Net == nil {
		return errors.New("no net")
	}

	d.Server = amqp.NewServer(conn, d.Net)
	return nil
}

func (d *Device) Run(ctx context.Context) error {
	return d.Server.Serve(ctx)
}
