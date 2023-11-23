package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/sequence"
	amqpGo "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"sync"
	"sync/atomic"
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
	Marking  control.Marking
}

type Controller struct {
	logger          *zap.Logger
	ch              *amqpGo.Channel
	mu              sync.Mutex
	discoveryCtx    context.Context
	discoveryCancel context.CancelFunc
	cmd             *amqp.CommandService
	event           *amqp.EventService
	dataCh          chan *control.Event
	q               *amqpGo.Queue
	Routes          map[string]*Instance
	CurrentStep     *atomic.Int32
	Sequence        *sequence.Sequence
	Net             *labeled.Net
	dataRelay       chan *control.Event
	Known           map[string]map[string]*Instance
	exchange        string
	stepCh          chan struct{}
}

type WaitFor struct {
	From string
	Name string
}

func (c *Controller) Close() {
	c.discoveryCancel()
}

func (c *Controller) ActualMarking() control.Marking {
	ret := make(control.Marking)
	for _, v := range c.Known {
		for _, i := range v {
			for k, v := range i.Marking {
				ret[k] = v
			}
		}
	}
	return ret
}

func (c *Controller) ChannelData(ch chan *control.Event) {
	c.dataRelay = ch
}

func (c *Controller) DeviceMarking() map[string]control.Marking {
	ret := make(map[string]control.Marking)
	for devID, instance := range c.Routes {
		ret[devID] = make(control.Marking)
		ret[devID] = instance.Marking
	}
	return ret
}

func (c *Controller) MarkingIs(marking control.Marking) bool {
	actual := c.ActualMarking()
	for k, v := range marking {
		if actual[k] != v {
			return false
		}
	}
	return true
}

func (c *Controller) DevicesReady() error {
	if !c.MarkingIs(c.Sequence.InitialMarking) {
		return errors.New("initial Marking incorrect")
	}
	return nil
}

func (c *Controller) ValidSequence() error {
	if c.Sequence == nil {
		return errors.New("no sequence")
	}
	if !labeled.ValidSequence(c.Sequence.Net, c.Sequence.Events()) {
		return errors.New("invalid sequence")
	}
	return nil
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

func NewController(logger *zap.Logger, ch *amqpGo.Channel, exchange string) *Controller {
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
		logger:      logger,
		ch:          ch,
		q:           &q,
		cmd:         &amqp.CommandService{},
		event:       &amqp.EventService{},
		exchange:    exchange,
		CurrentStep: new(atomic.Int32),
		Routes:      make(map[string]*Instance),
		Known:       make(map[string]map[string]*Instance),
	}
	c.runDiscoverLoop(context.Background())
	return c
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

func (c *Controller) CloseRelay() {
	close(c.dataRelay)
	c.dataRelay = nil
}

func channelSteps(stepQueue []*sequence.Step) <-chan *sequence.Step {
	ch := make(chan *sequence.Step)
	go func() {
		for _, step := range stepQueue {
			ch <- step
		}
		close(ch)
	}()
	return ch
}

func (c *Controller) Start(ctx context.Context) {
	c.logger.Info("Starting sequence", zap.String("sequence", c.Sequence.Name))
	stepCh := channelSteps(c.Sequence.Steps)
	c.CurrentStep.Store(0)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				step, ok := <-stepCh
				if !ok {
					c.logger.Info("Sequence complete", zap.String("sequence", c.Sequence.Name))
					c.CurrentStep.Store(0)
					return
				}
				c.logger.Info("Starting step", zap.String("step", step.Name))
				inst := c.Routes[step.Device.ID]
				if inst == nil {
					log.Fatalf("Instance for device %s not found", step.Device.ID)
				}
				err := c.startStep(ctx, step)
				if err != nil {
					log.Println(err)
				}
				data := <-c.dataCh
				c.logger.Info("Received event", zap.String("event", data.Name))
				if c.dataRelay != nil {
					c.dataRelay <- data
				}
				if step.Action.Event.ID != data.Event.ID {
					log.Fatalf("Expected event %s (id: %s) but got %s (id: %s)", step.Action.Event.Name, step.Action.Event.ID, data.Event.Name, data.Event.ID)
				}
				c.CurrentStep.Add(1)
			}
		}
	}()
}

func (c *Controller) startStep(ctx context.Context, step *sequence.Step) error {
	to, found := c.Routes[step.Device.ID]
	if !found {
		return errors.New("device not found")
	}
	cmd := step.Command(to.ID)
	fmt.Printf("Sending %s (id: %s) to %s with data %v\n", cmd.Name, cmd.ID, cmd.To, cmd.Data)
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

func (c *Controller) registerInstance(deviceID, instanceID string, marking control.Marking) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Known[deviceID] == nil {
		c.logger.Debug("Registering device", zap.String("device", deviceID))
		c.Known[deviceID] = make(map[string]*Instance)
		c.logger.Debug("Registering instance", zap.String("device", deviceID), zap.String("instance", instanceID))
		c.Known[deviceID][instanceID] = &Instance{
			ID:       instanceID,
			liveness: MaxLiveness,
			Marking:  marking,
		}
		return
	}
	if c.Known[deviceID][instanceID] != nil {
		c.Known[deviceID][instanceID].liveness = MaxLiveness
		c.Known[deviceID][instanceID].Marking = marking
		return
	}
	c.logger.Debug("Registering instance", zap.String("device", deviceID), zap.String("instance", instanceID))
	c.Known[deviceID][instanceID] = &Instance{
		ID:       instanceID,
		liveness: MaxLiveness,
		Marking:  marking,
	}
	log.Printf("Registering instance %s for device %s with Marking %v", instanceID, deviceID, marking)
}

func (c *Controller) pruneInstances() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, vv := range c.Known {
		for instanceKey, v := range vv {
			if v == nil {
				continue
			}
			v.liveness--
			if v.liveness <= 0 {
				log.Printf("Pruning instance %s for device %s", v.ID, k)
				c.Known[k][instanceKey] = nil
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
		defer func() {
			close(c.dataCh)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case d := <-msgs:
				data, err := c.event.Load(ctx, d)
				if err != nil {
					log.Println(err)
					if err.Error() == "invalid routing key" {
						log.Println(data)
						panic(err)
					}
					continue
				}
				if data.Topic == "device" {
					c.registerInstance(data.Name, data.From, data.Marking)
					continue
				}
				fmt.Printf("Received %s from %s\n", data.Name, data.From)
				c.dataCh <- data
			}
		}
	}()
}
