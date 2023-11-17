package autosampler

import (
	"context"
	"embed"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

//go:embed device.yaml
var deviceYaml embed.FS

func Run(ctx context.Context, conn *amqp.Connection) {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	d := NewAutosampler(nil, nil)
	dev := d.Load()
	// any additional initialization goes here

	srv := server.New(dev.Nets[0], conn.Channel, Environ.Exchange, Environ.DeviceID, Environ.InstanceID, dev.EventMap(), d.Handlers(), logger)

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
