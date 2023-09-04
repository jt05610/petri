package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/binary"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/comm/serial"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"os/signal"
)

//go:embed device.yaml
var deviceYaml embed.FS

func receiveUntilNewLine(ch <-chan io.Reader) []byte {
	b := <-ch
	msg, err := io.ReadAll(b)
	if err != nil {
		log.Fatal(err)
	}
	return msg
}

func main() {
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
	msg := receiveUntilNewLine(rxCh)
	fmt.Printf("received: %s", msg)
	buf := new(bytes.Buffer)
	buf.WriteString("R")
	for i := 0; i < 3; i++ {

	}

	for i := 0; i < 8; i++ {
		// each needs to be two bites little endian
		err := binary.Write(buf, binary.LittleEndian, uint16(100))
		if err != nil {
			log.Fatal(err)
		}
	}
	buf.WriteString("\n")
	msg = buf.Bytes()
	// print the bytes of the message we are sending
	fmt.Printf("sending: %x\n", msg)

	if len(msg) != 18 {
		log.Fatal("Invalid length")
	}
	txCh <- msg
	go printChan(rxCh)
	if err != nil {
		logger.Fatal("Failed to open channel port", zap.Error(err))
	}
	d := NewMixingValve()
	dev := d.load()
	srv := server.New(dev.Nets[0], ch, exchange, deviceID, instanceID, dev.EventMap(), d.Handlers())
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

func printChan(ch <-chan io.Reader) {
	for {
		b := <-ch
		msg, err := io.ReadAll(b)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("received: %s", msg)
	}
}
