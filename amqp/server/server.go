package server

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp2 "github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Server struct {
	*labeled.Net
	logger     *zap.Logger
	name       string
	deviceName string
	ch         *amqp.Channel
	q          *amqp.Queue
	devEvents  <-chan *labeled.Event
	cmd        *amqp2.CommandService
	event      *amqp2.EventService
	netCh      <-chan *labeled.Event
	handlers   control.Handlers
	exchange   string
	deviceID   string
	instanceID string
}

func (s *Server) AddHandler(route string, f labeled.Handler) {
	s.handlers[route] = f
}

func (s *Server) commandName(routingKey string) string {
	return strings.Split(routingKey, ".")[2]
}

func (s *Server) route(data *control.Command) (*control.Event, error) {
	var res *control.Event
	s.logger.Info("Routing command", zap.String("command", data.Event.Name))
	done := make(chan struct{})
	go func() {
		defer close(done)
		ev := <-s.devEvents
		res = &control.Event{
			Event:   ev,
			Topic:   "event",
			Marking: make(map[string]int),
			From:    s.instanceID,
		}
	}()
	if err := s.Handle(context.Background(), data.Event); err != nil {
		s.logger.Error("Failed to handle event", zap.Error(err))
		return &control.Event{
			Event:   data.Event,
			Topic:   "error",
			From:    s.instanceID,
			Marking: s.MarkingMap(),
		}, err
	}
	<-done
	s.logger.Info("Handled event", zap.String("event", data.Event.Name))
	for k, v := range s.MarkingMap() {
		res.Marking[k] = v
	}
	return res, nil
}

func New(net *labeled.Net, ch *amqp.Channel, exchange string, deviceID string, instanceID string, eventMap map[string]*petri.Transition, handlers control.Handlers, logger *zap.Logger) *Server {
	for ev, h := range handlers {
		err := net.AddHandler(ev, eventMap[ev], h)
		failOnError(err, "Failed to add handler")
	}

	queues := make([]string, 0, len(handlers)+2)
	err := ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")
	for key := range handlers {
		queues = append(queues, instanceID+".commands."+key)
	}
	queues = append(queues, instanceID+".state.get")
	queues = append(queues, "devices")
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	for _, key := range queues {
		err := ch.QueueBind(
			q.Name,   // queue name
			key,      // routing key
			exchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}

	return &Server{
		Net:        net,
		ch:         ch,
		q:          &q,
		devEvents:  net.Channel(),
		exchange:   exchange,
		handlers:   handlers,
		deviceID:   deviceID,
		instanceID: instanceID,
		event:      &amqp2.EventService{},
		cmd:        &amqp2.CommandService{},
		logger:     logger,
	}
}

func (s *Server) publishBeacon() {
	event := &labeled.Event{
		Name: "info",
		Data: map[string]interface{}{
			"device_name":   s.deviceName,
			"instance_name": s.name,
		},
	}
	resp, err := s.cmd.Flush(context.Background(), event, s.MarkingMap())
	failOnError(err, "Failed to flush command")
	err = s.ch.PublishWithContext(
		context.Background(),
		s.exchange,
		s.instanceID+".device."+s.deviceID,
		false,
		false,
		resp,
	)
	failOnError(err, "Failed to publish beacon")
}

func (s *Server) Listen(ctx context.Context) {
	msgs, err := s.ch.Consume(
		s.q.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	failOnError(err, "Failed to register a consumer")
	s.netCh = s.Net.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Closing connection")
				return
			case <-s.netCh:
				fmt.Println("Received a message from the net")
			case d := <-msgs:
				s.logger.Debug("Received message", zap.String("routing_key", d.RoutingKey))
				if d.RoutingKey == "devices" {
					s.publishBeacon()
					continue
				}
				data, err := s.cmd.Load(context.Background(), d)
				failOnError(err, "Failed to load command")
				switch data.Topic {
				case "state":
					resp, err := s.cmd.Flush(context.Background(), data.Event, s.MarkingMap())
					failOnError(err, "Failed to flush command")
					err = s.ch.PublishWithContext(ctx, s.exchange, s.instanceID+".state.current", false, false, resp)
				case "commands":
					event, err := s.route(data)
					failOnError(err, "Failed to handle data")
					resp, err := s.cmd.Flush(ctx, event.Event, s.MarkingMap())
					log.Printf("Sending response %v", resp)
					err = s.ch.PublishWithContext(ctx,
						s.exchange,
						event.RoutingKey(),
						false,
						false,
						resp,
					)
					failOnError(err, "Failed to publish event response")
				}
			}
		}
	}()
	<-ctx.Done()
}
