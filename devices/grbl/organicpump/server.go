package main

import (
	"context"
	"embed"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/env"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"strconv"
)

//go:embed device.yaml
var deviceYaml embed.FS

func rpcClient(e *env.Environment) (proto.GRBLClient, error) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(e.RPCAddress, opts...)
	failOnError(err, "Failed to connect to RabbitMQ")
	return proto.NewGRBLClient(conn), nil
}

func loadPumpParams() *InitializeRequest {
	diam := os.Getenv("SYRINGE_DIAMETER")
	vol := os.Getenv("SYRINGE_VOLUME")
	diamFloat, err := strconv.ParseFloat(diam, 64)
	failOnError(err, "Failed to parse syringe diameter")
	volFloat, err := strconv.ParseFloat(vol, 64)
	failOnError(err, "Failed to parse syringe volume")
	return &InitializeRequest{
		SyringeDiameter: diamFloat,
		SyringeVolume:   volFloat,
	}
}

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
	client, err := rpcClient(environ)
	failOnError(err, "Failed to connect to GRBL")
	d := NewOrganicPump(client)
	req := loadPumpParams()
	dev := d.load()
	maxPos := os.Getenv("MAX_POS")
	maxPosFloat, err := strconv.ParseFloat(maxPos, 64)
	failOnError(err, "Failed to parse max pos")
	d.MaxPos = float32(maxPosFloat)
	_, err = d.Initialize(context.Background(), req)
	failOnError(err, "Failed to initialize device")
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

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
