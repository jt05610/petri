package aqueouspump

import (
	"context"
	"embed"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

//go:embed device.yaml
var deviceYaml embed.FS

func loadPumpParams() *InitializeRequest {
	return &InitializeRequest{
		SyringeDiameter: SyringeDiameter,
		SyringeVolume:   SyringeVolume,
	}
}

func Run(ctx context.Context, conn *amqp.Connection, client proto.GRBLServer) {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	d := NewAqueousPump(client)
	dev := d.load()
	req := loadPumpParams()
	_, err = d.Initialize(context.Background(), req)
	pos := float32(0)
	spd := float32(1000)
	failOnError(err, "Failed to initialize device")
	d.MaxPos = MaxPos
	_, err = d.Move(context.Background(), &proto.MoveRequest{
		Y:     &pos,
		Speed: &spd,
	})
	if err != nil {
		log.Fatal(err)
	}
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