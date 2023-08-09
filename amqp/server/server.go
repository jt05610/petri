package server

import (
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strings"
	"time"
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

type Handler struct {
	ch       *amqp.Channel
	q        *amqp.Queue
	handlers Handlers
	exchange string
	deviceID string
}

func (a *Handler) AddHandler(route string, f HandlerFunc) {
	a.handlers[route] = f
}

func (a *Handler) commandName(routingKey string) string {
	return strings.Split(routingKey, ".")[2]
}

func (a *Handler) Load(_ context.Context, data amqp.Delivery) (*Command, error) {
	command := a.commandName(data.RoutingKey)
	res := &Command{
		command: command,
	}
	return res, json.Unmarshal(data.Body, &res.data)
}

func (a *Handler) Flush(_ context.Context, event *Event) (amqp.Publishing, error) {
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

func (a *Handler) route(data *Command) (*Event, error) {
	return a.handlers.Handle(context.Background(), data)
}

func NewHandler(ch *amqp.Channel, exchange string, deviceID string, handlers Handlers) *Handler {
	queues := make([]string, 0, len(handlers))
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
	for _, key := range queues {
		err := ch.QueueBind(
			q.Name,   // queue name
			key,      // routing key
			exchange, // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}
	return &Handler{
		ch:       ch,
		q:        &q,
		exchange: exchange,
		handlers: handlers,
		deviceID: deviceID,
	}
}

func (a *Handler) Listen(ctx context.Context) {
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
				data, err := a.Load(context.Background(), d)
				failOnError(err, "Failed to load data")
				log.Printf("Received a message: %s", data)
				event, err := a.handlers.Handle(context.Background(), data)
				failOnError(err, "Failed to handle data")
				log.Printf("Handled message %s and got %s", data, event)
				resp, err := a.Flush(context.Background(), event)
				err = a.ch.PublishWithContext(context.Background(),
					a.exchange,
					event.routingKey(a.deviceID),
					false,
					false,
					resp,
				)
				failOnError(err, "Failed to publish a message")
			}
		}
	}()
}

func main() {
	err := godotenv.Load()

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

	failOnError(err, "Failed to declare an exchange")

	handlers := Handlers{
		"open_a": func(ctx context.Context, data *Command) (*Event, error) {
			return &Event{Name: data.command, Data: data.data}, nil
		},
		"open_b": func(ctx context.Context, data *Command) (*Event, error) {
			return &Event{Name: data.command, Data: data.data}, nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := NewHandler(ch, exchange, devID, handlers)
	handler.Listen(ctx)
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()
	<-ctx.Done()
}
