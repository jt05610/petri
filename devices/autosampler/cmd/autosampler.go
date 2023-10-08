package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/devices/autosampler"
	proto "github.com/jt05610/petri/devices/autosampler/proto"
	"github.com/jt05610/petri/env"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/signal"
)

func grpcConnect(environment *env.Environment) proto.AutosamplerClient {
	conn, err := grpc.Dial(environment.RPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return proto.NewAutosamplerClient(conn)
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	err = godotenv.Load()
	if err != nil {
		logger.Fatal("Failed to load .env", zap.Error(err))
	}
	environ := &autosampler.Environ
	conn, err := amqp.Dial(environ)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.Fatal("Failed to close connection", zap.Error(err))
		}
	}()
	client := grpcConnect(environ)
	d := autosampler.NewAutosampler(client)
	dev := d.Load()
	srv := server.New(dev.Nets[0], conn.Channel, environ.Exchange, environ.DeviceID, environ.InstanceID, dev.EventMap(), d.Handlers(), logger)
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
