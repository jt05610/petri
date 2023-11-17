package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp"
	"github.com/jt05610/petri/comm/serial"
	"github.com/jt05610/petri/devices/grbl/aqueouspump"
	"github.com/jt05610/petri/devices/grbl/organicpump"
	"github.com/jt05610/petri/devices/grbl/rheogrande"
	"github.com/jt05610/petri/devices/grbl/rheoten"
	"github.com/jt05610/petri/devices/grbl/twvalve"
	"github.com/jt05610/petri/env"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"github.com/jt05610/petri/grbl/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os/signal"

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
	RabbitmqUri      = "amqp://jrt:GJWJLOABBoFnrE8xiUnSONfMLBzWn7m@SOP-4470A-1/petri"
	AqPumpDeviceID   = ""
	AqPumpInstanceID = "clm9r7bd60000jgw4qvlgo3zb"
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
	_, err = s.Home(ctx, &proto.HomeRequest{})
	s.TxChan <- []byte("G55\n")
	if err != nil {
		logger.Fatal("Failed to home", zap.Error(err))
	}
	for !s.Cts.Load() {

	}
	resp, err := s.FloodOn(ctx, &proto.FloodOnRequest{})
	logger.Info("Flood on", zap.Any("resp", resp))
	// lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", environ.Port))
	//if err != nil {
	//	logger.Fatal("Failed to listen", zap.Error(err))
	//}
	opts := make([]grpc.ServerOption, 0)
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterGRBLServer(grpcServer, s)
	logger.Info("Starting grpc server", zap.Int("port", environ.Port))
	go func() {
		//err := grpcServer.Serve(lis)
		if err != nil {
			logger.Fatal("Failed to serve grpc", zap.Error(err))
		}
	}()
	connections := make([]*amqp.Connection, 5)
	for i := 0; i < 5; i++ {
		conn, err := amqp.Dial(amqpEnv())
		if err != nil {
			logger.Fatal("Failed to dial amqp", zap.Error(err))
		}
		connections[i] = conn
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	defer func() {
		for _, conn := range connections {
			err := conn.Close()
			if err != nil {
				logger.Error("Failed to close amqp connection", zap.Error(err))
			}
		}
	}()
	go organicpump.Run(ctx, connections[0], s)
	go aqueouspump.Run(ctx, connections[1], s)
	go rheogrande.Run(ctx, connections[2], s)
	go rheoten.Run(ctx, connections[3], s)
	go twvalve.Run(ctx, connections[4], s)

	<-ctx.Done()
}
