package client

import (
	"context"
	"encoding/json"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strings"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Command struct {
	To string
	*labeled.Event
}

func (c *Command) snakeCaseName() string {
	return strings.Replace(strings.ToLower(c.Name), " ", "_", -1)
}

func (c *Command) routingKey() string {
	return c.To + ".commands." + c.snakeCaseName()
}

type Event struct {
	From string
	*labeled.Event
}

type Controller struct {
	ch       *amqp.Channel
	dataCh   chan *Event
	q        *amqp.Queue
	routes   map[string]string
	exchange string
}

func (a *Controller) Load(_ context.Context, data amqp.Delivery) (*Event, error) {
	sk := strings.Split(data.RoutingKey, ".")
	from := sk[0]
	command := sk[2]
	res := &Event{
		From: from,
		Event: &labeled.Event{
			Name: command,
		},
	}
	return res, json.Unmarshal(data.Body, &res.Event.Data)
}

func (a *Controller) Flush(_ context.Context, event *labeled.Event) (amqp.Publishing, error) {
	bytes, err := json.Marshal(event.Data)
	if err != nil {
		var zero amqp.Publishing
		return zero, err
	}

	return amqp.Publishing{
		Body:         bytes,
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Headers: amqp.Table{
			"x-event-name": event.Name,
		},
	}, nil
}

func NewController(ch *amqp.Channel, exchange string, routes map[string]string) *Controller {
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
	err = ch.QueueBind(
		q.Name,       // queue name
		"*.events.*", // routing key
		exchange,     // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")
	return &Controller{
		ch:       ch,
		q:        &q,
		exchange: exchange,
		routes:   routes,
	}
}

func (a *Controller) Send(ctx context.Context, cmd *Command) error {
	p, err := a.Flush(ctx, cmd.Event)
	if err != nil {
		return err
	}
	return a.ch.PublishWithContext(
		ctx,
		a.exchange,       // exchange
		cmd.routingKey(), // routing key
		false,            // mandatory
		false,            // immediate
		p,
	)
}

func (a *Controller) Start(ctx context.Context, step *db.StepModel) error {
	to, found := a.routes[step.Action().Device().ID]
	if !found {
		return nil
	}
	cmd := &Command{
		To: to,
		Event: &labeled.Event{
			Name: step.Action().Event().Name,
			Data: step.Action().Constants(),
		},
	}
	ctx, timeout := context.WithTimeout(ctx, time.Duration(1)*time.Second)
	defer timeout()

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

func (a *Controller) Data() <-chan *Event {
	return a.dataCh
}

func (a *Controller) Listen(ctx context.Context) {
	a.dataCh = make(chan *Event)
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
				data, err := a.Load(ctx, d)
				if err != nil {
					log.Println(err)
					continue
				}
				a.dataCh <- data
			}
		}
	}()
}
