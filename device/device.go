package device

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"io"
	"log/slog"
	"strings"
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

type ConnectionKind int

const (
	RemotePlaceLocalTransition ConnectionKind = iota
	LocalPlaceRemoteTransition
	RemotePlaceRemoteTransition
	LocalPlaceLocalTransition
)

type RemoteConnection struct {
	Kind ConnectionKind
	amqp.ArcDirection
	Exchange string
	Addr     string
}

func linksNets(a *petri.Arc) bool {
	sourceNet := strings.Split(a.Src.Identifier(), ".")[0]
	destNet := strings.Split(a.Dest.Identifier(), ".")[0]
	return sourceNet != destNet
}

func (d *Device) remoteConnections() ([]*RemoteConnection, error) {
	ret := make([]*RemoteConnection, 0)
	for exchange, addr := range d.Remotes {
		subNet := d.Subnet(exchange)
		if subNet == nil {
			return nil, errors.New("no subnet")
		}
		for _, arc := range d.SubnetArcs(exchange) {
			if !linksNets(arc) {
				continue
			}
			c := &RemoteConnection{
				Exchange: exchange,
				Addr:     addr,
				Kind:     d.describeArc(arc),
			}
			fmt.Printf("connecting arc: %v = %v\n", arc, c)
			ret = append(ret, c)
		}
	}
	return ret, nil
}

func (d *Device) isRemote(n petri.Node) bool {
	pl := d.Place(n.Identifier())
	if pl == nil {
		return false
	}
	split := strings.Split(pl.ID, ".")
	if len(split) == 1 {
		return false
	}
	netName := split[0]
	_, found := d.Remotes[netName]
	return found
}

func (d *Device) describeArc(a *petri.Arc) ConnectionKind {
	if d.isRemote(a.Src) {
		if d.isRemote(a.Dest) {
			return RemotePlaceRemoteTransition
		}
		return RemotePlaceLocalTransition
	}
	if d.isRemote(a.Dest) {
		return LocalPlaceRemoteTransition
	}
	return LocalPlaceLocalTransition
}

func (d *Device) arcDirection(pl *petri.Place, tr *petri.Transition) amqp.ArcDirection {
	placeToTrans := d.Net.Arc(pl, tr)
	transToPlace := d.Net.Arc(tr, pl)
	if placeToTrans != nil && transToPlace != nil {
		return amqp.ToAndFromPlace
	}
	if placeToTrans != nil {
		return amqp.ToPlace
	}
	if transToPlace != nil {
		return amqp.FromPlace
	}
	return amqp.None
}

func ConnectGRPCNet[T any](ctx context.Context, url string, clientFactory func(connInterface grpc.ClientConnInterface) T) (T, error) {
	conn, err := grpc.Dial(url)
	if err != nil {
		var zero T
		return zero, err
	}
	go func() {
		<-ctx.Done()
		err := conn.Close()
		if err != nil {
			slog.Error("error closing connection", slog.Any("message", err))
		}
	}()
	client := clientFactory(conn)
	return client, nil
}

func (d *Device) Connect(ctx context.Context, url string) error {
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
	d.Server = amqp.NewServer(conn, d.Net, m.Mark())
	for rem, netUrl := range d.Remotes {
		net := d.Subnet(rem)
		if net == nil {
			return errors.New("no subnet")
		}
		err := d.Server.ConnectPlaces(net, netUrl)
		if err != nil {
			return err
		}
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
