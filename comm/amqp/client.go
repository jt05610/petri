package amqp

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Codecs struct {
	Connect    RPCCodec[ConnectRequest, Response]
	Disconnect RPCCodec[DisconnectRequest, Response]
}

type Client struct {
	ch *amqp.Channel
	amqp.Queue
	codecs   Codecs
	timeout  time.Duration
	Exchange string
	msgs     <-chan amqp.Delivery
}

func NewClient(conn *amqp.Connection, exchange string, timeout time.Duration) (*Client, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}
	err = ch.QueueBind("", q.Name, exchange, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		ch: ch,
		codecs: Codecs{
			Connect:    ConnectCodec,
			Disconnect: DisconnectCodec,
		},
		Queue:    q,
		timeout:  timeout,
		Exchange: exchange,
		msgs:     msgs,
	}, nil
}

func (c *Client) listen(ctx context.Context) <-chan Response {

	out := make(chan Response)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-c.msgs:
				var response Response
				err := c.codecs.Disconnect.Response.Unmarshal(m.Body, &response)
				if err != nil {
					panic(err)
				}
				out <- response
			}
		}
	}()
	return out
}

func (c *Client) Close() {
	err := c.ch.Close()
	if err != nil {
		return
	}
}

func (c *Client) Connect(ctx context.Context, request *ConnectRequest) error {
	return roundTrip(ctx, c, c.codecs.Connect.Request, "connect", request)
}

func (c *Client) Disconnect(ctx context.Context, request *DisconnectRequest) error {
	return roundTrip(ctx, c, c.codecs.Disconnect.Request, "disconnect", request)
}

func roundTrip[T any](ctx context.Context, c *Client, codec Codec[T], path string, req *T) error {
	var err error
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp := c.listen(ctx)
	b, err := codec.Marshal(req)
	if err != nil {
		return err
	}
	fmt.Printf("publishing to exchange %s with routing key %s\n", c.Exchange, path)
	err = c.ch.PublishWithContext(ctx, c.Exchange, path, false, false, amqp.Publishing{
		Body:          b,
		CorrelationId: petri.ID(),
		ReplyTo:       c.Name,
	})
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case r := <-resp:
		return errors.New(r.Error)
	}
}
