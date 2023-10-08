package main

import (
	"context"
	"embed"
	"errors"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/comm/serial"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
)

//go:embed device.yaml
var deviceYaml embed.FS

type RunConfig struct {
	URI        string `json:"uri"`
	Exchange   string `json:"exchange"`
	DeviceID   string `json:"device_id"`
	InstanceID string `json:"instance_id"`
}

var (
	URINotFound        = errors.New("uri not found")
	ExchangeNotFound   = errors.New("exchange not found")
	DeviceIDNotFound   = errors.New("device_id not found")
	InstanceIDNotFound = errors.New("instance_id not found")
)

type ConfigLookup struct {
	key   string
	value *string
	err   error
}

func (c *ConfigLookup) LookupEnv(key string) error {
	v, ok := os.LookupEnv(key)
	if !ok {
		return c.err
	}
	*c.value = v
	return nil
}

func LoadEnv() (config *RunConfig, err error) {
	err = godotenv.Load()
	failOnError(err, "Error loading .env file")
	cfg := new(RunConfig)
	lookups := []*ConfigLookup{
		{
			key:   "RABBITMQ_URI",
			value: &cfg.URI,
			err:   URINotFound,
		},
		{
			key:   "AMQP_EXCHANGE",
			value: &cfg.Exchange,
			err:   ExchangeNotFound,
		},
		{
			key:   "DEVICE_ID",
			value: &cfg.DeviceID,
			err:   DeviceIDNotFound,
		},
		{
			key:   "INSTANCE_ID",
			value: &cfg.InstanceID,
			err:   InstanceIDNotFound,
		},
	}
	for _, lookup := range lookups {
		err := lookup.LookupEnv(lookup.key)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func Run(ctx context.Context, d *MixingValve, logger *zap.Logger, cfg *RunConfig) {
	conn, err := amqp.Dial(cfg.URI)
	failOnError(err, "Failed to connect to RabbitMQ")
	logger.Info("Connected to RabbitMQ", zap.String("uri", cfg.URI))
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
	port, err := serial.OpenPort("COM6", 115200)
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
	_ = receiveUntilNewLine(rxCh)
	*d = *NewMixingValve(txCh, rxCh)
	dev := d.load()
	srv := server.New(dev.Nets[0], ch, cfg.Exchange, cfg.DeviceID, cfg.InstanceID, dev.EventMap(), d.Handlers(), logger)
	logger.Info("Started 🐰 server")
	srv.Listen(ctx)
	<-ctx.Done()
}

func main() {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	logger.Info("Starting 🐰 server")
	// Setup rabbitmq channel
	cfg, err := LoadEnv()
	failOnError(err, "Error loading config")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	d := new(MixingValve)
	go Run(ctx, d, logger, cfg)
	<-c // Wait for SIGINT
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func receiveUntilNewLine(ch <-chan io.Reader) (msg []byte) {
	var err error
	for {
		b := <-ch
		msg, err = io.ReadAll(b)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(string(msg), "ready") {
			return msg
		}
	}
}
