package queue

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

// Queue is a wrapper for amqp.Queue that implements
type Queue struct {
	amqp.Queue
	Schema   *petri.TokenSchema
	ch       *amqp.Channel
	exchange string
}

func NewQueue(ch *amqp.Channel, schema *petri.TokenSchema, exchange, name string) *Queue {
	err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	q, err := ch.QueueDeclare(
		name,
		true,
		false,

		false,
		true,
		nil,
	)
	if err != nil {
		panic(err)
	}
	err = ch.QueueBind(
		"",
		name,
		exchange,
		false,
		nil,
	)
	if err != nil {
		panic(err)

	}
	return &Queue{
		Queue:    q,
		exchange: exchange,
		ch:       ch,
		Schema:   schema,
	}
}

// put notifies everything that cares about this place that a token has been added. This does not actually add a token to the queue, as this is done by a separate operation
func (q *Queue) put(ctx context.Context, t *petri.Token) error {
	return q.publish(ctx, t, q.Name+".in")
}

// pop notifies everything that cares about this place that a token has been removed. This does not actually remove a token from the queue, as this is done by a separate operation
func (q *Queue) pop(ctx context.Context, t *petri.Token) error {
	return q.publish(ctx, t, q.Name+".out")
}

func Message(t *petri.Token) amqp.Publishing {
	return amqp.Publishing{
		ContentType: "text/plain",
		Timestamp:   time.Now(),
		Body:        t.Bytes(),
	}
}

func (q *Queue) publish(ctx context.Context, t *petri.Token, route string) error {
	return q.ch.PublishWithContext(
		ctx,
		q.exchange,
		route,
		false,
		false,
		Message(t),
	)
}

func (q *Queue) subscribe(ctx context.Context, route string) (<-chan amqp.Delivery, error) {
	return q.ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (q *Queue) get(ctx context.Context, route string) (*petri.Token, error) {
	queue, err := q.ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	msg := RPCMessage(&queue)
	err = q.ch.PublishWithContext(
		ctx,
		q.exchange,
		route,
		false,
		false,
		msg,
	)
	if err != nil {
		return nil, err
	}
	msgs, err := q.ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case received := <-msgs:
			if received.CorrelationId == msg.CorrelationId {
				return q.Schema.NewToken(received.Body)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

func (q *Queue) rpc(ctx context.Context, route string, tt ...*petri.Token) (<-chan *petri.Token, error) {
	queue, err := q.ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	err = q.ch.QueueBind("", queue.Name, q.exchange, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := q.ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	msg := RPCMessage(&queue, tt...)

	err = q.ch.PublishWithContext(
		ctx,
		q.exchange,
		q.Name+"."+route,
		false,
		false,
		msg,
	)
	if err != nil {
		return nil, err
	}
	ch := make(chan *petri.Token)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				panic(ctx.Err())
			case m := <-msgs:
				if m.CorrelationId != msg.CorrelationId {
					continue
				}
				t, err := q.Schema.NewToken(m.Body)
				if err != nil {
					fmt.Println(err)
				}
				if m.Headers["empty"].(bool) {
					return
				}
				ch <- t
				if m.Headers["done"].(bool) {
					return
				}
			}
		}
	}()
	return ch, nil

}

func (q *Queue) post(ctx context.Context, route string, t *petri.Token) error {
	queue, err := q.ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	msg := RPCMessage(&queue, t)
	err = q.ch.PublishWithContext(
		ctx,
		q.exchange,
		route,
		false,
		false,
		msg,
	)
	if err != nil {
		return err
	}
	msgs, err := q.ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	for {
		select {
		case m := <-msgs:
			if m.CorrelationId == msg.CorrelationId {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func RPCMessage(q *amqp.Queue, t ...*petri.Token) amqp.Publishing {
	if len(t) == 0 {
		return amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: petri.ID(),
			ReplyTo:       q.Name,
			Timestamp:     time.Now(),
		}
	}
	if len(t) > 1 {
		panic("too many tokens")
	}
	return amqp.Publishing{
		ContentType:   "text/plain",
		CorrelationId: petri.ID(),
		Timestamp:     time.Now(),
		ReplyTo:       q.Name,
		Body:          t[0].Bytes(),
	}
}
