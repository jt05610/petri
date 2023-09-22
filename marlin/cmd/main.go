package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/comm/serial"
	fracCollector "github.com/jt05610/petri/devices/fraction_collector"
	"github.com/jt05610/petri/marlin"
	proto "github.com/jt05610/petri/marlin/proto/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
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
	ctx := context.Background()
	s := marlin.New(ctx, port, logger)

	go func() {
		err := s.Listen(ctx)
		if err != nil {
			panic(err)
		}
	}()
	go s.RunHeartbeat(ctx)
	_, err = s.Home(ctx, &proto.HomeRequest{})
	if err != nil {
		logger.Fatal("Failed to home", zap.Error(err))
	}
	if err != nil {
		logger.Fatal("Failed to move", zap.Error(err))
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", environ.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	opts := make([]grpc.ServerOption, 0)
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterMarlinServer(grpcServer, s)
	logger.Info("Starting grpc server", zap.Int("port", environ.Port))
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			logger.Fatal("Failed to serve grpc", zap.Error(err))
		}
	}()
	go fracCollector.Run(ctx, s)
	<-ctx.Done()
}
