package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp"
	proto "github.com/jt05610/petri/comm/grbl/proto/v1"
	"github.com/jt05610/petri/comm/grbl/server"
	"github.com/jt05610/petri/comm/serial"
	"github.com/jt05610/petri/devices/grbl/pump_bank"
	"github.com/jt05610/petri/env"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os/signal"
	"time"

	"os"
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

const (
	RabbitmqUri = "amqp://jrt:GJWJLOABBoFnrE8xiUnSONfMLBzWn7m@SOP-4470A-1/petri"
)

type Environment struct {
	Port         int
	SerialPort   string
	Baud         int
	StartupBlock string
}

func load() *Environment {
	vals := map[string]*string{
		"SERIAL_PORT":   new(string),
		"SERIAL_BAUD":   new(string),
		"STARTUP_BLOCK": new(string),
		"PORT":          new(string),
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
	}
}

func amqpEnv() *env.Environment {
	return &env.Environment{
		URI:      RabbitmqUri,
		Exchange: "topic_devices",
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

	s := server.New(port, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.RunHeartbeat(ctx)

	_, err = s.SpindleOff(ctx, &proto.SpindleOffRequest{})
	if err != nil {
		logger.Fatal("Failed to spindle off", zap.Error(err))
	}

	_, err = s.FloodOn(ctx, &proto.FloodOnRequest{})
	if err != nil {
		logger.Fatal("Failed to flood on", zap.Error(err))
	}
	time.Sleep(1 * time.Second)
	_, err = s.CoolantOff(ctx, &proto.CoolantOffRequest{})
	if err != nil {
		logger.Fatal("Failed to flood off", zap.Error(err))
	}
	_, err = s.SpindleOff(ctx, &proto.SpindleOffRequest{})
	if err != nil {
		logger.Fatal("Failed to spindle on", zap.Error(err))
	}

	_, err = s.Home(ctx, &proto.HomeRequest{})
	s.TxChan <- []byte("G55\n")
	if err != nil {
		logger.Fatal("Failed to home", zap.Error(err))
	}
	for !s.Cts.Load() {
	}
	// lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", environ.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	opts := make([]grpc.ServerOption, 0)
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterGRBLServer(grpcServer, s)
	logger.Info("Starting grpc server", zap.Int("port", environ.Port))
	//go func() {
	//err := grpcServer.Serve(lis)
	//if err != nil {
	//	logger.Fatal("Failed to serve grpc", zap.Error(err))
	//	}
	//}()
	conn, err := amqp.Dial(amqpEnv())
	if err != nil {
		logger.Fatal("Failed to dial amqp", zap.Error(err))
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.Error("Failed to close amqp connection", zap.Error(err))
		}
	}()

	go pump_bank.Run(ctx, conn, s)
	<-ctx.Done()
}
