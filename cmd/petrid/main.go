package main

import (
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Command struct {
	command string
	data    interface{}
}

func (c *Command) routingKey(devID string) string {
	return devID + ".commands." + c.command
}

type EventDataService interface {
	Load(ctx context.Context, data amqp.Delivery) (*Command, error)
	Flush(ctx context.Context, command *Command) (amqp.Publishing, error)
}

type Event struct {
	Name string
	Data interface{}
}

func (e *Event) WithData(data interface{}) *Event {
	e.Data = data
	return e
}

func (e *Event) routingKey(devID string) string {
	return devID + ".events." + e.Name
}

type amqpEventDataService struct {
	ch              *amqp.Channel
	handlers        Handlers
	deviceID        string
	commandEventMap map[string]*Event
}

func (a *amqpEventDataService) commandName(routingKey string) string {
	return strings.Split(routingKey, ".")[2]
}

func (a *amqpEventDataService) Load(_ context.Context, data amqp.Delivery) (*Command, error) {
	command := a.commandName(data.RoutingKey)
	res := &Command{
		command: command,
	}
	return res, json.Unmarshal(data.Body, &res.data)
}

func (a *amqpEventDataService) Flush(_ context.Context, event *Event) (amqp.Publishing, error) {
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

type HandlerFunc func(ctx context.Context, data *Command) (*Event, error)

type Handlers map[string]HandlerFunc

func (h Handlers) Handle(ctx context.Context, data *Command) (*Event, error) {
	return (h)[data.command](ctx, data)
}

func (a *amqpEventDataService) route(data *Command) (*Event, error) {
	return a.handlers.Handle(context.Background(), data)
}

func main() {
	err := godotenv.Load()
	srv := amqpEventDataService{
		commandEventMap: map[string]*Event{
			"open_a": {
				Name: "open_a",
			},
			"open_b": {
				Name: "open_b",
			},
		},
	}
	srv.handlers = Handlers{
		"open_a": func(ctx context.Context, data *Command) (*Event, error) {
			return srv.commandEventMap[data.command].WithData(data.data), nil
		},
		"open_b": func(ctx context.Context, data *Command) (*Event, error) {
			return srv.commandEventMap[data.command].WithData(data.data), nil
		},
	}

	failOnError(err, "Error loading .env file")
	uri := os.Getenv("RABBITMQ_URI")
	devID := os.Getenv("DEVICE_ID")
	exchange := "topic_devices"
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer func() {
		err := conn.Close()
		failOnError(err, "Failed to close connection to RabbitMQ")
	}()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer func() {
		err := ch.Close()
		failOnError(err, "Failed to close channel")
	}()
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
	srv.ch = ch
	failOnError(err, "Failed to declare a queue")

	for _, key := range []string{devID + ".commands.open_a", devID + ".commands.open_b"} {
		err = ch.QueueBind(
			q.Name,   // queue name
			key,      // routing key
			exchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	done := make(chan struct{})
	go func() {
		for d := range msgs {
			data, err := srv.Load(context.Background(), d)
			failOnError(err, "Failed to load data")
			log.Printf("Received a message: %s", data)
			event, err := srv.handlers.Handle(context.Background(), data)
			failOnError(err, "Failed to handle data")
			log.Printf("Handled message %s and got %s", data, event)
			resp, err := srv.Flush(context.Background(), event)
			err = ch.PublishWithContext(context.Background(),
				exchange,
				event.routingKey(devID),
				false,
				false,
				resp,
			)
			failOnError(err, "Failed to publish a message")
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-done
}
