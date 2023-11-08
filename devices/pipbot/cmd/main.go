package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/comm/serial"
	fracCollector "github.com/jt05610/petri/devices/PipBot"
	"github.com/jt05610/petri/env"
	"github.com/jt05610/petri/marlin"
	proto "github.com/jt05610/petri/marlin/proto/v1"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strconv"
)

func getVal(name string, s *string) error {
	v, found := os.LookupEnv(name)
	if !found {
		return fmt.Errorf("environment variable %s not found", name)
	}
	*s = v
	return nil
}

type Environment struct {
	Port         int
	SerialPort   string
	Baud         int
	StartupBlock string
	RabbitMqURI  string
	AmqpExchange string
}

func load() *Environment {
	vals := map[string]*string{
		"SERIAL_PORT":   new(string),
		"SERIAL_BAUD":   new(string),
		"STARTUP_BLOCK": new(string),
		"PORT":          new(string),
		"RABBITMQ_URI":  new(string),
		"AMQP_EXCHANGE": new(string),
	}
	for k, v := range vals {
		err := getVal(k, v)
		if err != nil {
			panic(err)
		}
	}
	baud, err := strconv.ParseInt(*vals["SERIAL_BAUD"], 10, 64)
	port, err := strconv.ParseInt(*vals["PORT"], 10, 64)
	if err != nil {
		panic(err)
	}
	return &Environment{
		SerialPort:   *vals["SERIAL_PORT"],
		Baud:         int(baud),
		StartupBlock: *vals["STARTUP_BLOCK"],
		Port:         int(port),
		RabbitMqURI:  *vals["RABBITMQ_URI"],
		AmqpExchange: *vals["AMQP_EXCHANGE"],
	}
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
	environ := load()
	port, err := serial.OpenPort(environ.SerialPort, environ.Baud)
	if err != nil {
		logger.Fatal("Failed to open port", zap.Error(err))
	}
	defer func() {
		err := port.Close()
		if err != nil {
			logger.Error("Failed to close port", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	s := marlin.New(ctx, port, logger)
	go func() {
		err := s.Listen(ctx)
		if err != nil {
			panic(err)
		}
	}()
	initLines := [][]byte{
		[]byte("M301 S1\n"),
		[]byte("M82\n"),
		[]byte("G92 E0\n"),
		[]byte("G92 E-40\n"),
		[]byte("G1 F500 E0\n"),
		[]byte("M400\n"),
		[]byte("G92 E0\n"),
		[]byte("M82\n"),
	}
	go s.RunHeartbeat(ctx)
	_, err = s.Home(ctx, &proto.HomeRequest{})
	for _, line := range initLines {
		_, err := port.WritePort(line)
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}
	conn, err := amqp.Dial(&env.Environment{URI: environ.RabbitMqURI, Exchange: environ.AmqpExchange})

	go fracCollector.Run(fracCollector.Autosampler, ctx, conn, s)
	<-ctx.Done()
}
