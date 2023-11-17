package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/comm/serial"
	"github.com/jt05610/petri/devices/autosampler"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strconv"
)

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
	port, found := os.LookupEnv("PORT")
	if !found {
		panic("PORT not found")
	}

	baud, found := os.LookupEnv("BAUD")
	if !found {
		panic("BAUD not found")
	}
	baudInt, err := strconv.Atoi(baud)
	if err != nil {
		panic(err)
	}
	ser, err := serial.OpenPort(port, baudInt)
	if err != nil {
		panic(err)
	}
	d := autosampler.NewAutosampler(ser, logger)
	for _, c := range autosampler.InitializeCommands() {
		err = d.Do(context.Background(), c)
		if err != nil {
			panic(err)
		}
	}
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
	nInj := 3
	for i := 0; i < nInj; i++ {
		req := &autosampler.InjectRequest{
			InjectionVolume: 10,
			Position:        "J1",
			AirCushion:      10,
			ExcessVolume:    0,
			FlushVolume:     50,
			NeedleDepth:     -10,
		}
		_, err := d.Inject(ctx, req)
		if err != nil {
			panic(err)
		}
		_, err = d.Injected(ctx, &autosampler.InjectedRequest{})
		logger.Info("Injected", zap.Int("i", i))
		if err != nil {
			panic(err)
		}
		_, err = d.WaitForReady(ctx, &autosampler.WaitForReadyRequest{})
		logger.Info("Ready", zap.Int("i", i))
		if err != nil {
			panic(err)
		}
	}
	srv.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}
