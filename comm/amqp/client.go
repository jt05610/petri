package amqp

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Codecs struct {
	Connect    RPCCodec[ConnectRequest, Response]
	Disconnect RPCCodec[DisconnectRequest, Response]
	Put        RPCCodec[PutRequest, Response]
	Pop        RPCCodec[PopRequest, PopResponse]
}

type PutRequest struct {
	Place string
	Token []byte
}

type PopRequest struct {
	Place string
}

type PopResponse struct {
	Place string
	Token petri.Token
	*Response
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
			Put:        PutCodec,
			Pop:        PopCodec,
		},
		Queue:    q,
		timeout:  timeout,
		Exchange: exchange,
		msgs:     msgs,
	}, nil
}

func listen[T any](ctx context.Context, c *Client, codec Codec[T]) <-chan T {
	out := make(chan T)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-c.msgs:
				var response T
				err := codec.Unmarshal(m.Body, &response)
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
	_, err := roundTrip(ctx, c, c.codecs.Connect, "connect", request)
	return err
}

func (c *Client) Disconnect(ctx context.Context, request *DisconnectRequest) error {
	_, err := roundTrip(ctx, c, c.codecs.Disconnect, "disconnect", request)
	return err
}

func (c *Client) Put(ctx context.Context, request *PutRequest) error {
	res, err := roundTrip(ctx, c, c.codecs.Put, "put", request)
	if err != nil {
		return err
	}
	if res.Error != "" {
		return fmt.Errorf(res.Error)
	}
	return nil
}

func (c *Client) Pop(ctx context.Context, request *PopRequest) (*PopResponse, error) {
	res, err := roundTrip(ctx, c, c.codecs.Pop, "pop", request)
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, fmt.Errorf(res.Error)
	}
	return res, nil
}

func roundTrip[T, U any](ctx context.Context, c *Client, codec RPCCodec[T, U], path string, req *T) (*U, error) {
	var err error
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp := listen(ctx, c, codec.Response)
	b, err := codec.Request.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("publishing to exchange %s with routing key %s\n", c.Exchange, path)
	err = c.ch.PublishWithContext(ctx, c.Exchange, path, false, false, amqp.Publishing{
		Body:          b,
		CorrelationId: petri.ID(),
		ReplyTo:       c.Name,
	})
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case r := <-resp:
		return &r, nil
	}
}
