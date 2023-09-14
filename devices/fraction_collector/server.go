package fracCollector

import (
	"context"
	"embed"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

//go:embed device.yaml
var deviceYaml embed.FS

func Run(ctx context.Context, srv marlin.MarlinServer) {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	err = godotenv.Load()
	failOnError(err, "Error loading .env file")

	logger.Info("Starting 🐰 server")
	// Setup rabbitmq channel
	uri, ok := os.LookupEnv("RABBITMQ_URI")
	if !ok {
		logger.Fatal("RABBITMQ_URI not set")
	}
	exchange, ok := os.LookupEnv("AMQP_EXCHANGE")
	if !ok {
		logger.Fatal("AMQP_EXCHANGE not set")
	}
	deviceID, ok := os.LookupEnv("DEVICE_ID")
	if !ok {
		logger.Fatal("DEVICE_ID not set")
	}
	instanceID, ok := os.LookupEnv("INSTANCE_ID")
	if !ok {
		logger.Fatal("INSTANCE_ID not set")
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

	d := NewFractionCollector(srv, pipbot.MakeGrid(31.5-29, 40-17))
	dev := d.load()
	logger, err = zap.NewProduction()
	failOnError(err, "Error creating logger")
	s := server.New(dev.Nets[0], ch, exchange, deviceID, instanceID, dev.EventMap(), d.Handlers(), logger)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	logger.Info("Started 🐰 server")
	s.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
