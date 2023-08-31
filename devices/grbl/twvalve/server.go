package main

import (
	"context"
	"embed"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/comm/serial"
	"github.com/jt05610/petri/env"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"time"
)

//go:embed device.yaml
var deviceYaml embed.FS

func main() {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	environ := env.LoadEnv(logger)
	conn, err := amqp.Dial(environ)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer func() {
		err := conn.Close()
		failOnError(err, "Failed to close connection")
	}()
	port, err := serial.OpenPort(environ.SerialPort, environ.Baud)
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
	srv := server.New(dev.Nets[0], conn.Channel, environ.Exchange, environ.DeviceID, environ.InstanceID, dev.EventMap(), d.Handlers())
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

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
