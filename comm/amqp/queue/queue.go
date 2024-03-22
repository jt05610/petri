package queue

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"strconv"
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
	fanout := exchange + ".marking"
	err = ch.ExchangeDeclare(fanout, "fanout", true, false, false, false, nil)
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
	return &Queue{
		Queue:    q,
		exchange: exchange,
		ch:       ch,
		Schema:   schema,
	}
}

// put notifies everything that cares about this place that a token has been added. This does not actually add a token to the queue, as this is done by a separate operation
func (q *Queue) put(ctx context.Context, t petri.Token) error {
	return q.publish(ctx, t, "in")
}

// pop notifies everything that cares about this place that a token has been removed. This does not actually remove a token from the queue, as this is done by a separate operation
func (q *Queue) pop(ctx context.Context, t petri.Token) error {
	return q.publish(ctx, t, "out")
}

func Message(t *petri.Token) amqp.Publishing {
	return amqp.Publishing{
		ContentType: "text/plain",
		Timestamp:   time.Now(),
		Body:        t.Bytes(),
	}
}

func PubMessage(t petri.Token, sequence int, kind string) amqp.Publishing {
	return amqp.Publishing{
		ContentType: "text/plain",
		Timestamp:   time.Now(),
		Body:        t.Bytes(),
		MessageId:   strconv.Itoa(sequence),
		Headers: map[string]interface{}{
			"kind": kind,
		},
	}
}

func (q *Queue) publish(ctx context.Context, t petri.Token, route string) error {
	msg := PubMessage(t, time.Now().Nanosecond(), route)
	return q.ch.PublishWithContext(
		ctx,
		q.exchange+".marking",
		"",
		false,
		false,
		msg,
	)
}

type update struct {
	op string
	petri.Token
	sequence int
}

func (q *Queue) subscribe(ctx context.Context) (<-chan update, error) {
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
	err = q.ch.QueueBind(queue.Name, "", q.exchange+".marking", false, nil)
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
	ch := make(chan update)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-msgs:
				if m.Headers["kind"] == nil {
					continue
				}
				t, err := q.Schema.NewToken(m.Body)
				if err != nil {
					fmt.Println(err)
				}
				upd := update{
					op:    m.Headers["kind"].(string),
					Token: t,
				}
				fmt.Println(m.MessageId, m.Headers, string(m.Body), upd)
				ch <- upd
			}
		}
	}()
	return ch, nil
}

func (q *Queue) get(ctx context.Context, route string) (petri.Token, error) {
	queue, err := q.ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return petri.Token{}, err
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
		return petri.Token{}, err
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
		return petri.Token{}, err
	}
	for {
		select {
		case received := <-msgs:
			if received.CorrelationId == msg.CorrelationId {
				return q.Schema.NewToken(received.Body)
			}
		case <-ctx.Done():
			return petri.Token{}, ctx.Err()
		}
	}
}

// rpc is a helper function for making rpc calls to the queue
func (q *Queue) rpc(ctx context.Context, route string, tt ...petri.Token) (<-chan petri.Token, error) {
	// TODO: this maybe could be a generic with better typing, or have a schema passed to it. Or maybe it's fine as is.
	// TODO: decide if this is fine as is
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
	ch := make(chan petri.Token)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				slog.Error("rpc context done")
				return
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

func (q *Queue) post(ctx context.Context, route string, t petri.Token) error {
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

func RPCMessage(q *amqp.Queue, t ...petri.Token) amqp.Publishing {
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
