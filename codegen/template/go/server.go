package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"os"
)

type Valve struct {
	*server.Server
}

func (v *Valve) OpenA(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *Valve) OpenB(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *Valve) Handlers() control.Handlers {
	return control.Handlers{
		"open_a": v.OpenA,
		"open_b": v.OpenB,
	}
}

func main() {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	err = godotenv.Load()
	failOnError(err, "Error loading .env file")

	logger.Info("Starting 🐰 server")
	// Setup rabbitmq channel
	uri := os.Getenv("AMQP_URI")
	if uri == "" {
		logger.Fatal("AMQP_URI not set")
	}
	exchange, ok := os.LookupEnv("AMQP_EXCHANGE")
	if !ok {
		logger.Fatal("AMQP_EXCHANGE not set")
	}
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")
	logger.Info("Connected to RabbitMQ", zap.String("uri", uri))
	defer func() {
		err := conn.Close()
		failOnError(err, "Failed to close connection")
	}()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	logger.Info("Opened channel")
	defer func() {
		err := ch.Close()
		failOnError(err, "Failed to close channel")
	}()

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
