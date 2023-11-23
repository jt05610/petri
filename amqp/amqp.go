package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/env"
	"github.com/jt05610/petri/labeled"
	amqp "github.com/rabbitmq/amqp091-go"
	"strings"
)

type CommandService struct{}

func (a *CommandService) Load(_ context.Context, data amqp.Delivery) (*control.Command, error) {
	sk := strings.Split(data.RoutingKey, ".")
	to := sk[0]
	topic := sk[1]
	command := sk[2]
	id := ""
	if data.Headers != nil {
		if data.Headers["x-event-id"] != nil {
			id = data.Headers["x-event-id"].(string)
		}
	}
	res := &control.Command{
		To:    to,
		Topic: topic,
		Event: &labeled.Event{
			ID:   id,
			Name: command,
		},
	}
	if len(data.Body) == 0 {
		return res, nil
	}
	return res, json.Unmarshal(data.Body, &res.Event.Data)
}

func (a *CommandService) Flush(_ context.Context, event *labeled.Event, mark ...control.Marking) (amqp.Publishing, error) {
	bytes, err := json.Marshal(&event.Data)
	if err != nil {
		var zero amqp.Publishing
		return zero, err
	}
	headers := amqp.Table{
		"x-event-name": event.Name,
		"x-event-id":   event.ID,
	}
	if len(mark) > 0 {
		m := mark[0]
		mbytes, err := json.Marshal(&m)
		if err != nil {
			return amqp.Publishing{}, err
		}
		headers["x-marking"] = mbytes
	}

	return amqp.Publishing{
		Body:         bytes,
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
	}, nil
}

type EventService struct{}

func (a *EventService) Load(_ context.Context, data amqp.Delivery) (*control.Event, error) {
	sk := strings.Split(data.RoutingKey, ".")
	if len(sk) != 3 {
		return nil, errors.New("invalid routing key")
	}
	from := sk[0]
	topic := sk[1]
	event := sk[2]
	id := ""
	if data.Headers != nil {
		if data.Headers["x-event-id"] != nil {
			id = data.Headers["x-event-id"].(string)
		}
		if data.Headers["x-marking"] != nil {
			var m control.Marking
			if err := json.Unmarshal(data.Headers["x-marking"].([]byte), &m); err != nil {
				return nil, err
			}
			ret := &control.Event{
				From:    from,
				Topic:   topic,
				Event:   &labeled.Event{ID: id, Name: event},
				Marking: m,
			}
			return ret, json.Unmarshal(data.Body, &ret.Data)
		}

	}
	res := &control.Event{
		From:  from,
		Topic: topic,
		Event: &labeled.Event{
			Name: event,
			ID:   id,
		},
	}
	if len(data.Body) == 0 {
		return res, nil
	}
	return res, json.Unmarshal(data.Body, &res.Data)
}

type Connection struct {
	*amqp.Connection
	*amqp.Channel
}

func (c *Connection) Close() error {
	if c.Channel != nil {
		err := c.Channel.Close()
		if err != nil {
			return err
		}
	}
	return c.Connection.Close()
}

func Dial(environ *env.Environment) (*Connection, error) {
	conn, err := amqp.Dial(environ.URI)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &Connection{conn, ch}, nil
}
