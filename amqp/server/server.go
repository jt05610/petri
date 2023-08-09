package server

import (
	"context"
	"fmt"
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
	ch       *amqp.Channel
	q        *amqp.Queue
	cmd      *amqp2.CommandService
	event    *amqp2.EventService
	handlers control.Handlers
	exchange string
	deviceID string
}

func (s *Server) AddHandler(route string, f labeled.Handler) {
	s.handlers[route] = f
}

func (s *Server) commandName(routingKey string) string {
	return strings.Split(routingKey, ".")[2]
}

func (s *Server) route(data *control.Command) (*control.Event, error) {
	return s.handlers.Handle(context.Background(), data)
}

func New(ch *amqp.Channel, exchange string, deviceID string, handlers control.Handlers) *Server {

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
		fmt.Println("Binding queue", q.Name, "to exchange", exchange, "with routing key", key, "...")
		err := ch.QueueBind(
			q.Name,   // queue name
			key,      // routing key
			exchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}
	return &Server{
		ch:       ch,
		q:        &q,
		exchange: exchange,
		handlers: handlers,
		deviceID: deviceID,
		event:    &amqp2.EventService{},
		cmd:      &amqp2.CommandService{},
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
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Closing connection")
				return
			case d := <-msgs:
				log.Printf("Received a message: %s", d.Body)
				data, err := s.cmd.Load(context.Background(), d)
				failOnError(err, "Failed to load command")
				switch data.Topic {
				case "state":
					resp, err := s.cmd.Flush(context.Background(), data.Event)
					failOnError(err, "Failed to flush command")
					err = s.ch.PublishWithContext(ctx, s.exchange, s.deviceID+".state.current", false, false, resp)
				case "commands":
					log.Printf("Received command: %s", data)
					event, err := s.handlers.Handle(ctx, data)
					failOnError(err, "Failed to handle data")
					log.Printf("Handled message %s and got %s", data, event)
					resp, err := s.cmd.Flush(ctx, event.Event)
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
