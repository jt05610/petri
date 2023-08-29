package main

import (
	"context"
	"embed"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/server"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"
)

//go:embed device.yaml
var deviceYaml embed.FS

type Environment struct {
	URI        string
	Exchange   string
	DeviceID   string
	InstanceID string
	SerialPort string
	Baud       int
}

func LoadEnv(logger *zap.Logger) *Environment {
	err := godotenv.Load()
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
	serPort, found := os.LookupEnv("SERIAL_PORT")
	if !found {
		logger.Fatal("SERIAL_PORT not set")
	}
	baud, found := os.LookupEnv("SERIAL_BAUD")
	if !found {
		logger.Fatal("SERIAL_BAUD not set")
	}
	baudInt, err := strconv.ParseInt(baud, 10, 64)
	if err != nil {
		logger.Fatal("Failed to parse baud", zap.Error(err))
	}
	return &Environment{
		URI:        uri,
		Exchange:   exchange,
		DeviceID:   deviceID,
		InstanceID: instanceID,
		SerialPort: serPort,
		Baud:       int(baudInt),
	}
}

func main() {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	env := LoadEnv(logger)
	conn, err := amqp.Dial(env.URI)
	failOnError(err, "Failed to connect to RabbitMQ")
	logger.Info("Connected to RabbitMQ", zap.String("uri", env.URI))
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

	port, err := OpenPort(env.SerialPort, env.Baud)
	defer func() {
		err := port.Close()
		failOnError(err, "Failed to close port")
	}()
	if err != nil {
		logger.Fatal("Failed to open port", zap.Error(err))
	}
	txCh := make(chan []byte, 100)
	if err != nil {
		logger.Fatal("Failed to set read timeout", zap.Error(err))
	}
	rxCh, err := port.ChannelPort(context.Background(), txCh)
	<-rxCh
	if err != nil {
		logger.Fatal("Failed to open channel port", zap.Error(err))
	}

	go runHeartbeat(context.Background(), txCh)
	d := NewTwoPositionThreeWayValve(txCh, rxCh)
	go func() {
		err := d.Listen(context.Background())
		if err != nil {
			logger.Fatal("Failed to listen", zap.Error(err))
		}
	}()
	txCh <- []byte("$X\n")
	dev := d.load()
	srv := server.New(dev.Nets[0], ch, env.Exchange, env.DeviceID, env.InstanceID, dev.EventMap(), d.Handlers())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	logger.Info("Started 🐰 server")
	srv.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func doHeartbeat(txCh chan []byte) {
	txCh <- []byte("?\n")
}

func runHeartbeat(ctx context.Context, txCh chan []byte) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			doHeartbeat(txCh)
		}
	}
}
