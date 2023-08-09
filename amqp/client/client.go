package client

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma/db"
	amqpGo "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Controller struct {
	ch       *amqpGo.Channel
	cmd      *amqp.CommandService
	event    *amqp.EventService
	dataCh   chan *control.Event
	q        *amqpGo.Queue
	Routes   map[string]string
	exchange string
}

func NewController(ch *amqpGo.Channel, exchange string, routes map[string]string) *Controller {
	err := ch.Confirm(false)
	failOnError(err, "Failed to set confirm mode")
	err = ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	topics := []string{"*.events.*", "*.state.current"}
	for _, t := range topics {
		err = ch.QueueBind(
			q.Name,   // queue name
			t,        // routing key
			exchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind queue")
	}

	return &Controller{
		ch:       ch,
		q:        &q,
		cmd:      &amqp.CommandService{},
		event:    &amqp.EventService{},
		exchange: exchange,
		Routes:   routes,
	}
}

func (a *Controller) sendPing(ctx context.Context, deviceID string) error {
	return a.ch.PublishWithContext(
		ctx,
		a.exchange,            // exchange
		deviceID+".state.get", // routing key
		false,                 // mandatory
		false,                 // immediate
		amqpGo.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqpGo.Persistent,
			Body:         []byte{},
		},
	)
}

func (a *Controller) Ping(ctx context.Context, deviceID string) (bool, error) {
	retries := 3
	for i := 0; i < retries; i++ {
		err := a.sendPing(ctx, deviceID)
		if err != nil {
			return false, err
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(time.Duration(1) * time.Second):
			continue
		case recv := <-a.dataCh:
			if recv.From != deviceID {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}

func (a *Controller) Send(ctx context.Context, cmd *control.Command) error {
	p, err := a.cmd.Flush(ctx, cmd.Event)
	if err != nil {
		return err
	}
	fmt.Printf("sending %v to %s\n", cmd, cmd.RoutingKey())
	return a.ch.PublishWithContext(
		ctx,
		a.exchange,       // exchange
		cmd.RoutingKey(), // routing key
		false,            // mandatory
		false,            // immediate
		p,
	)
}

func (a *Controller) Start(ctx context.Context, step *db.StepModel, data interface{}) error {
	to, found := a.Routes[step.Action().Device().ID]
	if !found {
		return nil
	}
	cmd := &control.Command{
		To: to,
		Event: &labeled.Event{
			Name: step.Action().Event().Name,
			Data: data,
		},
	}

	log.Printf("Sending %s to %s", cmd.Name, cmd.To)
	done := make(chan struct{})
	var sendErr error
	go func() {
		err := a.Send(ctx, cmd)
		if err != nil {
			sendErr = err
		}
		log.Printf("Sent %s to %s", cmd.Name, cmd.To)
		close(done)
	}()

	select {
	case <-done:
		return sendErr
	case <-ctx.Done():
		log.Printf("Timed out sending %s to %s", cmd.Name, cmd.To)
		return ctx.Err()
	}
}

func (a *Controller) Data() <-chan *control.Event {
	return a.dataCh
}

func (a *Controller) Listen(ctx context.Context) {
	a.dataCh = make(chan *control.Event)
	msgs, err := a.ch.Consume(
		a.q.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d := <-msgs:
				data, err := a.event.Load(ctx, d)
				if err != nil {
					log.Println(err)
					continue
				}
				a.dataCh <- data
			}
		}
	}()
}
