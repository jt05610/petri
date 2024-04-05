package queue

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"sync"
	"time"
)

var _ petri.TokenQueue = (*Public)(nil)

type Public struct {
	*Queue
	tokens   []petri.Token
	mu       sync.Mutex
	bindings map[string]func(context.Context, amqp.Delivery)
	updated  chan []petri.Token
}

func (q *Public) Close() {
	err := q.ch.Close()
	if err != nil {
		panic(err)
	}
}

func (q *Public) Peek(_ context.Context) ([]petri.Token, error) {
	return q.tokens, nil
}

func (q *Public) Enqueue(ctx context.Context, token ...petri.Token) error {
	hasUpdated := false
	for _, t := range token {
		q.tokens = append(q.tokens, t)
		err := q.Queue.put(ctx, t)
		if err != nil {
			return err
		}
		hasUpdated = true
	}
	if hasUpdated {
		fmt.Println("notifying")
		select {
		case q.updated <- q.tokens:
		default:
			slog.Info("local public queue not updated because the channel is not ready")
		}
		fmt.Println("notified")
	}
	return nil
}

func (q *Public) Dequeue(ctx context.Context) (petri.Token, error) {
	if len(q.tokens) == 0 {
		return petri.Token{}, nil
	}
	t := q.tokens[0]
	err := q.pop(ctx, t)
	if err != nil {
		return petri.Token{}, err
	}
	q.tokens = q.tokens[1:]
	select {
	case q.updated <- q.tokens:
	default:
		slog.Info("local public queue not updated because the channel is not ready")
	}
	return t, nil
}

func (q *Public) Copy() petri.TokenQueue {
	return &Public{
		Queue: q.Queue,
	}
}

func (q *Public) Monitor(ctx context.Context) <-chan []petri.Token {
	ch := make(chan []petri.Token)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case ch <- <-q.updated:
			}
		}
	}()
	return ch
}

func (q *Public) Available(ctx context.Context) (int, error) {
	return len(q.tokens), nil
}

func NewLocal(exchange string, conn *amqp.Connection, pl *petri.Place) *Public {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return &Public{
		Queue:   NewQueue(ch, pl.AcceptedTokens[0], exchange, pl.ID),
		updated: make(chan []petri.Token, 100),
		tokens:  make([]petri.Token, 0, 100),
	}
}

func (q *Public) handlePeek(ctx context.Context, msg amqp.Delivery) {
	tokens, err := q.Peek(ctx)
	if err != nil {
		panic(err)
	}
	n := len(tokens)
	for i, t := range tokens {
		err := q.ch.PublishWithContext(
			ctx,
			q.exchange,
			msg.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: msg.CorrelationId,
				Body:          t.Bytes(),
				Headers: map[string]interface{}{
					"empty": false,
					"done":  i == n-1,
				},
			},
		)
		if err != nil {
			panic(err)
		}
	}
}

func (q *Public) handleAvailable(ctx context.Context, msg amqp.Delivery) {
	n := len(q.tokens)
	err := q.ch.PublishWithContext(
		ctx,
		q.exchange,
		msg.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: msg.CorrelationId,
			Body:          petri.NewIntValue(n).Bytes(),
			Headers: map[string]interface{}{
				"empty": false,
				"done":  true,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

func (q *Public) handleGet(ctx context.Context, msg amqp.Delivery) {
	t, err := q.Dequeue(ctx)
	if err != nil {
		panic(err)
	}
	if t.Value == nil {
		err = q.ch.PublishWithContext(
			ctx,
			q.exchange,
			msg.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: msg.CorrelationId,
				Headers: map[string]interface{}{
					"empty": true,
					"done":  true,
				},
			},
		)
		if err != nil {
			panic(err)
		}
		return
	}
	err = q.ch.PublishWithContext(
		ctx,
		q.exchange,
		msg.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: msg.CorrelationId,
			Body:          t.Bytes(),
			Headers: map[string]interface{}{
				"empty": false,
				"done":  true,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

func (q *Public) handlePost(ctx context.Context, msg amqp.Delivery) {
	t, err := q.Schema.NewToken(msg.Body)
	if err != nil {
		panic(err)
	}
	err = q.Enqueue(ctx, t)
	if err != nil {
		panic(err)
	}
	err = q.post(ctx, msg.ReplyTo, t)
	if err != nil {
		panic(err)
	}
}

// Serve allows the local queue to handle RPCs from remote queues that want to interact with this one. Available RPCs are peek, get, and post.
func (q *Public) Serve(ctx context.Context) error {
	// declare the bindings and attach
	q.bindings = map[string]func(context.Context, amqp.Delivery){
		q.Name + ".peek":      q.handlePeek,
		q.Name + ".get":       q.handleGet,
		q.Name + ".post":      q.handlePost,
		q.Name + ".available": q.handleAvailable,
	}
	queue, err := q.ch.QueueDeclare("", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for b := range q.bindings {
		err := q.ch.QueueBind(queue.Name, b, q.exchange, false, nil)
		if err != nil {
			return err
		}
		slog.Info("queue bound", slog.String("binding", b), slog.String("exchange", q.exchange))
	}
	err = q.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	messages, err := q.ch.Consume(queue.Name, "", true, false, false, false, nil)

	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case m, more := <-messages:
			if !more {
				return nil
			}
			go q.process(ctx, m)
		}
	}
}

var Timeout = 1 * time.Second

func (q *Public) process(ctx context.Context, m amqp.Delivery) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	hndl, found := q.bindings[m.RoutingKey]
	if !found {
		if m.RoutingKey != "" {
			slog.Error("no handler for", slog.String("routing_key", m.RoutingKey))
		}
		return
	}
	hndl(ctx, m)
}
