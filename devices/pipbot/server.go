package PipBot

import (
	"context"
	"embed"
	"github.com/jt05610/petri/amqp"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

//go:embed device.yaml
var deviceYaml embed.FS

func Run(grid Grid, ctx context.Context, conn *amqp.Connection, client marlin.MarlinServer) {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	d := NewPipBot(grid, []int{3}, 12, client, logger)
	// dev := d.load()
	// any additional initialization goes here

	// srv := server.New(dev.Nets[0], conn.Channel, environ.Exchange, environ.DeviceID, environ.InstanceID, dev.EventMap(), d.Handlers(), logger)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		logger.Info("Started http server")
		err = runServer(d, ":8088")
		if err != nil {
			logger.Error("Error running server", zap.Error(err))
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	logger.Info("Started 🐰 server")

	// srv.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func runServer(d *PipBot, addr string) error {
	mux := d.Mux()

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return srv.ListenAndServe()
}
