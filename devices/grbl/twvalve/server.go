package twvalve

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

func Run(ctx context.Context, conn *amqp.Connection, client proto.GRBLServer) {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	d := NewTwoPositionThreeWayValve(client)
	dev := d.load()
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

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
