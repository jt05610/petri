package main

import (
	"context"
	"github.com/jt05610/petri/amqp"
	camera "github.com/jt05610/petri/devices/gocam"
	"github.com/jt05610/petri/env"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

func main() {
	logger, err := zap.NewProduction()
	camera.FailOnError(err, "Error creating logger")
	environ := env.LoadEnv(logger)
	conn, err := amqp.Dial(environ)
	camera.FailOnError(err, "Failed to connect to RabbitMQ")
	logger.Info("Connected to RabbitMQ", zap.String("uri", environ.URI))
	defer func() {
		err := conn.Close()
		camera.FailOnError(err, "Failed to close connection")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	logger.Info("Started 🐰 server")
	go camera.Run(ctx, conn)
	<-ctx.Done()
}
