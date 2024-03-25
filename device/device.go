package device

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp"
	"github.com/rabbitmq/amqp091-go"
	"io"
	"log/slog"
)

type Device struct {
	Net *petri.Net
	*amqp.Server
	Remotes     map[string]string
	Initial     map[string]string
	connections map[string]*amqp091.Connection
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
	d.connections = make(map[string]*amqp091.Connection)
	conn, err := amqp091.Dial(url)
	if err != nil {
		return err
	}
	d.connections[url] = conn
	if d.Net == nil {
		return errors.New("no net")
	}
	m := d.Net.NewMarking()
	for pl, mark := range d.Initial {
		place := d.Net.Place(pl)
		schema := place.AcceptedTokens[0]
		tok, err := schema.NewToken([]byte(mark))
		if err != nil {
			return err
		}
		err = m.PlaceTokens(place, tok)
		if err != nil {
			return err
		}
	}
	d.Server = amqp.NewServer(conn, d.Net, m)
	remotes := make(map[string]struct{})
	for exch, remote := range d.Remotes {
		var n *petri.Net
		for _, net := range d.Nets {
			if net.Name == exch {
				n = net
				break
			}
		}
		if n == nil {
			return fmt.Errorf("no net: %s", n)
		}
		conn, err := amqp091.Dial(remote)
		if err != nil {
			return err
		}
		d.connections[exch] = conn
		for _, pl := range n.Places {
			slog.Info("creating queue", slog.String("name", pl.Name), slog.String("type", "remote"))
			d.Server = d.Server.WithClientPlace(exch, conn, pl)
			remotes[pl.ID] = struct{}{}
		}
	}
	for _, pl := range d.Places {
		if _, ok := remotes[pl.ID]; ok {
			continue
		}
		if pl.IsEvent {
			continue
		}
		d.Server = d.Server.WithPublicPlaces(pl)
	}
	for _, a := range d.Arcs {
		if !a.LinksNets {
			continue
		}
		fmt.Printf("%v\n", a)
	}
	return nil
}

func (d *Device) Run(ctx context.Context) error {
	return d.Server.Serve(ctx)
}

func (d *Device) Close() {
	for _, conn := range d.connections {
		err := conn.Close()
		if err != nil {
			slog.Error("error closing connection", slog.Any("message", err))
		}
	}
}
