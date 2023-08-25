package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma/db"
	"github.com/jt05610/petri/sequence"
	amqpGo "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

const MaxLiveness = 3

type Instance struct {
	ID       string
	liveness int
	marking  control.Marking
}

type Controller struct {
	ch              *amqpGo.Channel
	mu              sync.Mutex
	discoveryCtx    context.Context
	discoveryCancel context.CancelFunc
	cmd             *amqp.CommandService
	event           *amqp.EventService
	dataCh          chan *control.Event
	q               *amqpGo.Queue
	Routes          map[string]*Instance
	Sequence        *sequence.Sequence
	Known           map[string]map[string]*Instance
	exchange        string
}

func (c *Controller) Close() {
	c.discoveryCancel()
}

func (c *Controller) Discover() error {
	return c.ch.PublishWithContext(
		c.discoveryCtx,
		c.exchange, // exchange
		"devices",  // routing key
		false,      // mandatory
		false,      // immediate
		amqpGo.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqpGo.Persistent,
			Body:         []byte{},
		},
	)
}

func (c *Controller) runDiscoverLoop(ctx context.Context) {
	c.discoveryCtx, c.discoveryCancel = context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-c.discoveryCtx.Done():
				return
			case <-time.After(time.Duration(1) * time.Second):
				c.pruneInstances()
				err := c.Discover()
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()
}

func NewController(ch *amqpGo.Channel, exchange string) *Controller {
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
	topics := []string{"*.events.*", "*.state.current", "*.device.*"}
	for _, t := range topics {
		err = ch.QueueBind(
			q.Name,   // queue name
			t,        // routing key
			exchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind queue")
	}
	c := &Controller{
		ch:       ch,
		q:        &q,
		cmd:      &amqp.CommandService{},
		event:    &amqp.EventService{},
		exchange: exchange,
		Routes:   make(map[string]*Instance),
		Known:    make(map[string]map[string]*Instance),
	}
	c.runDiscoverLoop(context.Background())
	return c
}

func (c *Controller) sendPing(ctx context.Context, deviceID string) error {
	return c.ch.PublishWithContext(
		ctx,
		c.exchange,            // exchange
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

func (c *Controller) Ping(ctx context.Context, deviceID string) (control.Marking, error) {
	retries := 3
	for i := 0; i < retries; i++ {
		err := c.sendPing(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(1) * time.Second):
			continue
		case recv := <-c.dataCh:
			if recv.From != deviceID {
				continue
			}
			return recv.Marking, nil
		}
	}
	return nil, errors.New("ping timed out")
}

func (c *Controller) Send(ctx context.Context, cmd *control.Command) error {
	p, err := c.cmd.Flush(ctx, cmd.Event)
	if err != nil {
		return err
	}
	fmt.Printf("sending %v to %s\n", cmd, cmd.RoutingKey())
	return c.ch.PublishWithContext(
		ctx,
		c.exchange,       // exchange
		cmd.RoutingKey(), // routing key
		false,            // mandatory
		false,            // immediate
		p,
	)
}

func (c *Controller) Start(ctx context.Context, step *db.StepModel, data map[string]interface{}) error {
	to, found := c.Routes[step.Action().Device().ID]
	if !found {
		return errors.New("device not found")
	}
	cmd := &control.Command{
		To: to.ID,
		Event: &labeled.Event{
			Name: step.Action().Event().Name,
			Data: data,
		},
	}

	done := make(chan struct{})
	var sendErr error
	go func() {
		err := c.Send(ctx, cmd)
		if err != nil {
			sendErr = err
		}
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

func (c *Controller) Data() <-chan *control.Event {
	return c.dataCh
}

func (c *Controller) registerInstance(deviceID, instanceID string, marking control.Marking) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Known[deviceID] == nil {
		c.Known[deviceID] = make(map[string]*Instance)
	}
	if c.Known[deviceID][instanceID] != nil {
		c.Known[deviceID][instanceID].liveness = MaxLiveness
		c.Known[deviceID][instanceID].marking = marking
		return
	}
	c.Known[deviceID][instanceID] = &Instance{
		ID:       instanceID,
		liveness: MaxLiveness,
		marking:  marking,
	}
	log.Printf("Registering instance %s for device %s with marking %v", instanceID, deviceID, marking)
}

func (c *Controller) pruneInstances() {
	for k, vv := range c.Known {
		for instanceKey, v := range vv {
			v.liveness--
			if v.liveness <= 0 {
				log.Printf("Pruning instance %s for device %s", v.ID, k)
				delete(c.Known[k], instanceKey)
			}
		}
	}
}

func (c *Controller) Listen(ctx context.Context) {
	c.dataCh = make(chan *control.Event)
	msgs, err := c.ch.Consume(
		c.q.Name, // queue
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
				data, err := c.event.Load(ctx, d)
				if err != nil {
					log.Println(err)
					continue
				}
				if data.Topic == "device" {
					go c.registerInstance(data.Name, data.From, data.Marking)
					continue
				}
				c.dataCh <- data
			}
		}
	}()
}
