package amqp

import (
	"context"
	"encoding/json"
	"github.com/jt05610/petri/control"
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
	res := &control.Command{
		To:    to,
		Topic: topic,
		Event: &labeled.Event{
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
	from := sk[0]
	topic := sk[1]
	event := sk[2]
	if data.Headers != nil {
		if data.Headers["x-marking"] != nil {
			var m control.Marking
			if err := json.Unmarshal(data.Headers["x-marking"].([]byte), &m); err != nil {
				return nil, err
			}
			ret := &control.Event{
				From:    from,
				Topic:   topic,
				Event:   &labeled.Event{Name: event},
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
		},
	}

	return res, json.Unmarshal(data.Body, &res.Data)
}
