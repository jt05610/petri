package camera

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
	FailOnError(err, "Error creating logger")
	d := NewCamera()
	dev := d.load()
	// any additional initialization goes here

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

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
