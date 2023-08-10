package server

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp2 "github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	amqp "github.com/rabbitmq/amqp091-go"
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
	ch        *amqp.Channel
	q         *amqp.Queue
	devEvents <-chan *labeled.Event
	cmd       *amqp2.CommandService
	event     *amqp2.EventService
	netCh     <-chan *labeled.Event
	handlers  control.Handlers
	exchange  string
	deviceID  string
}

func (s *Server) AddHandler(route string, f labeled.Handler) {
	s.handlers[route] = f
}

func (s *Server) commandName(routingKey string) string {
	return strings.Split(routingKey, ".")[2]
}

func (s *Server) route(data *control.Command) (*control.Event, error) {
	var res *control.Event
	fmt.Printf("Handling %s with current marking %v", data.Event.Name, s.MarkingMap())
	done := make(chan struct{})
	go func() {
		defer close(done)
		ev := <-s.devEvents
		res = &control.Event{
			Event:   ev,
			Topic:   "event",
			Marking: make(map[string]int),
			From:    "",
		}
	}()
	if err := s.Handle(context.Background(), data.Event); err != nil {
		log.Printf("Failed to handle event: %v", err)
		return &control.Event{
			Event:   data.Event,
			Topic:   "error",
			From:    "",
			Marking: s.MarkingMap(),
		}, err
	}
	<-done
	fmt.Printf("Handled %s with current marking %v", data.Event.Name, s.MarkingMap())
	for k, v := range s.MarkingMap() {
		res.Marking[k] = v
	}
	return res, nil
}

func New(net *labeled.Net, ch *amqp.Channel, exchange string, deviceID string, eventMap map[string]*petri.Transition, handlers control.Handlers) *Server {
	for ev, h := range handlers {
		err := net.AddHandler(ev, eventMap[ev], h)
		failOnError(err, "Failed to add handler")
	}

	queues := make([]string, 0, len(handlers)+1)
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
		queues = append(queues, deviceID+".commands."+key)
	}
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")
	queues = append(queues, deviceID+".state.get")
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
		Net:       net,
		ch:        ch,
		q:         &q,
		devEvents: net.Channel(),
		exchange:  exchange,
		handlers:  handlers,
		deviceID:  deviceID,
		event:     &amqp2.EventService{},
		cmd:       &amqp2.CommandService{},
	}
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
				log.Printf("Received a message: %s", d.Body)
				data, err := s.cmd.Load(context.Background(), d)
				failOnError(err, "Failed to load command")
				switch data.Topic {
				case "state":
					resp, err := s.cmd.Flush(context.Background(), data.Event, s.MarkingMap())
					failOnError(err, "Failed to flush command")
					err = s.ch.PublishWithContext(ctx, s.exchange, s.deviceID+".state.current", false, false, resp)
				case "commands":
					event, err := s.route(data)
					failOnError(err, "Failed to handle data")
					resp, err := s.cmd.Flush(ctx, event.Event, s.MarkingMap())
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
