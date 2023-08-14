package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/server"
	modbus "github.com/jt05610/petri/proto/v1"
	amqp "github.com/rabbitmq/amqp091-go"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
	"os/signal"
)

//go:embed device.yaml
var deviceYaml embed.FS

//go:embed secrets/cert.pem
var certFs embed.FS

func connect(serverAddr string) *grpc.ClientConn {
	certPool := x509.NewCertPool()
	caCert, err := certFs.ReadFile("secrets/cert.pem")
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(caCert)
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					CurvePreferences: []tls.CurveID{tls.CurveP256},
					MinVersion:       tls.VersionTLS12,
					RootCAs:          certPool,
				},
			),
		),
	)
	if err != nil {
		panic(err)
	}
	return conn
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
	deviceID, ok := os.LookupEnv("DEVICE_ID")
	if !ok {
		logger.Fatal("DEVICE_ID not set")
	}
	exchange, ok := os.LookupEnv("AMQP_EXCHANGE")
	if !ok {
		logger.Fatal("AMQP_EXCHANGE not set")
	}
	serverAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		logger.Fatal("SERVER_ADDR not set")
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
	rpcConn := connect(serverAddr)
	defer func() {
		err := rpcConn.Close()
		failOnError(err, "Failed to close connection")
	}()
	spClient := modbus.NewModbusClient(rpcConn)
	d := NewSyringePump(spClient)
	dev := d.load()
	srv := server.New(dev.Nets[0], ch, exchange, deviceID, dev.EventMap(), d.Handlers())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	logger.Info(fmt.Sprintf("🐰 server listening for commands at %s.commands.#", deviceID))
	srv.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
