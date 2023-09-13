package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/comm/serial"
	"github.com/jt05610/petri/marlin"
	proto "github.com/jt05610/petri/marlin/proto/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var _ proto.MarlinServer = (*Server)(nil)

type Server struct {
	*marlin.Parser
	cts           atomic.Bool
	logger        *zap.Logger
	machineStatus *atomic.Pointer[marlin.Status]
	state         *atomic.Pointer[proto.State]
	port          *serial.Port
	rxChan        <-chan io.Reader
	txChan        chan []byte
	listenCancel  context.CancelFunc
	proto.UnimplementedMarlinServer
}

func (s *Server) FanOn(ctx context.Context, request *proto.FanOnRequest) (*proto.Response, error) {
	err := s.do([]byte("M106 S255\n"), true, func(state *proto.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message: "ok",
	}, nil
}

func (s *Server) FanOff(ctx context.Context, request *proto.FanOffRequest) (*proto.Response, error) {
	err := s.do([]byte("M106 S0\n"), true, func(state *proto.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message: "ok",
	}, nil
}

func (s *Server) currentState() *proto.State {
	return s.state.Load()
}

func (s *Server) status() *marlin.Status {
	return s.machineStatus.Load()
}

func (s *Server) do(cmd []byte, synchronous bool, check func(state *proto.State) bool) error {
	for !s.cts.Load() {
	}
	s.txChan <- cmd
	s.cts.Store(false)
	if synchronous {
		for {
			status := s.currentState()
			if status == nil {
				continue
			}
			if s.cts.Load() && check(status) {
				return nil
			}
		}
	}
	return nil
}

func (s *Server) Home(ctx context.Context, req *proto.HomeRequest) (*proto.Response, error) {
	if req.Axis == nil {
		s.txChan <- []byte("G28\n")
	} else {
		if v, ok := req.Axis.(*proto.HomeRequest_X); ok {
			if v.X {
				s.txChan <- []byte("G28 X\n")
			}
		} else if v, ok := req.Axis.(*proto.HomeRequest_Y); ok {
			if v.Y {
				s.txChan <- []byte("G28 Y\n")
			}
		} else if v, ok := req.Axis.(*proto.HomeRequest_Z); ok {
			if v.Z {
				s.txChan <- []byte("G28 Z\n")
			}
		}
	}

	// wait until we see all positions at 0
	for {
		time.Sleep(100 * time.Millisecond)
		status := s.status()
		if status == nil {
			continue
		}
		if status.Position.X == 0 && status.Position.Y == 0 && status.Position.Z == 0 {
			return &proto.Response{
				Message: "ok",
			}, nil
		}
	}
}

func New(port *serial.Port, logger *zap.Logger) *Server {
	buf := new(bytes.Buffer)
	txCh := make(chan []byte, 100)
	rxCh, err := port.ChannelPort(context.Background(), txCh)
	if err != nil {
		panic(err)
	}
	ctx, can := context.WithCancel(context.Background())
	s := &Server{
		Parser:        marlin.NewParser(buf),
		port:          port,
		rxChan:        rxCh,
		cts:           atomic.Bool{},
		logger:        logger,
		state:         &atomic.Pointer[proto.State]{},
		machineStatus: &atomic.Pointer[marlin.Status]{},
		txChan:        txCh,
		listenCancel:  can,
	}
	go func() {
		err := s.Listen(ctx)
		if err != nil {
			panic(err)
		}
	}()
	return s
}

func (s *Server) Close() error {
	s.listenCancel()
	return s.port.Close()
}

func (s *Server) StateStream(req *proto.StateStreamRequest, server proto.Marlin_StateStreamServer) error {
	for {
		select {
		case <-server.Context().Done():
			return nil
		case <-time.After(1 * time.Second):
			state := s.currentState()
			if state == nil {
				continue
			}
			err := server.Send(&proto.StateStreamResponse{
				State:     state,
				Timestamp: time.Now().Format(time.RFC3339Nano),
			})
			if err != nil {
				return err
			}
		}
	}
}

func goToMsg(pos *proto.MoveRequest) []byte {
	bld := bytes.NewBuffer([]byte("G1"))
	if pos.Speed != nil {
		bld.WriteString(fmt.Sprintf(" F%.3f", *pos.Speed))
	}
	if pos.X != nil {
		bld.WriteString(fmt.Sprintf(" X%.3f", *pos.X))
	}
	if pos.Y != nil {
		bld.WriteString(fmt.Sprintf(" Y%.3f", *pos.Y))
	}
	if pos.Z != nil {
		bld.WriteString(fmt.Sprintf(" Z%.3f", *pos.Z))
	}
	if pos.E != nil {
		bld.WriteString(fmt.Sprintf(" E%.3f", *pos.E))
	}
	bld.WriteString("\nM400\n")
	ret := bld.Bytes()
	fmt.Println(string(ret))
	return ret
}

func (s *Server) Move(ctx context.Context, req *proto.MoveRequest) (*proto.Response, error) {
	err := s.do(goToMsg(req), true, func(state *proto.State) bool {
		if req.X != nil {
			if state.Position.X != *req.X {
				return false
			}
		}
		if req.Y != nil {
			if state.Position.Y != *req.Y {
				return false
			}
		}
		if req.Z != nil {
			if state.Position.Z != *req.Z {
				return false
			}
		}
		if req.E != nil {
			if state.Position.E != *req.E {
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message: "ok",
		Response: &proto.Response_Move{
			Move: &proto.MoveResponse{
				Message: "ok",
			},
		},
	}, nil
}

var (
	HeartbeatMsg = "M114 R\n"
)

func (s *Server) UpdateStatus(status marlin.StatusUpdate) {
	if _, ok := status.(*marlin.Ack); ok {
		s.cts.Store(true)
		return
	}
	if _, ok := status.(*marlin.Processing); ok {
		return
	}

	if upd, ok := status.(*marlin.Status); ok {
		s.logger.Debug("Received status update", zap.Any("status", upd))
		s.machineStatus.Store(upd)
		newState := &proto.State{
			Position: &proto.Position{
				X: upd.Position.X,
				Y: upd.Position.Y,
				Z: upd.Position.Z,
				E: upd.Position.E,
			},
		}
		s.state.Store(newState)
	}
}

func (s *Server) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-s.rxChan:
			bb, err := io.ReadAll(msg)
			if err != nil {
				s.logger.Error("Failed to read message", zap.Error(err))
				continue
			}
			s.logger.Debug("Received message", zap.String("msg", string(bb)))
			buf := bytes.NewBuffer(bb)
			parser := marlin.NewParser(buf)
			upd, err := parser.Parse()
			if err != nil {
				// write back to the buffer so that we can try again
				s.logger.Error("Failed to parse message", zap.Error(err))
			} else {
				s.UpdateStatus(upd)

			}
		}
	}
}

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

func (s *Server) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(300 * time.Millisecond)
	errCount := 3
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			errCount = 3
			if !s.cts.Load() {
				errCount--
				if errCount == 0 {
					s.logger.Fatal("Failed to receive ack")
				}
				continue
			}
			s.cts.Store(false)
			s.txChan <- []byte(HeartbeatMsg)
		}
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
	s := New(port, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.runHeartbeat(ctx)
	startupLines := strings.Split(environ.StartupBlock, "\n")
	for _, line := range startupLines {
		// wait for cts
		s.logger.Info("Sending line", zap.String("line", line))
		s.txChan <- []byte(line + "\n")
	}
	_, err = s.FanOn(ctx, &proto.FanOnRequest{})
	if err != nil {
		logger.Fatal("Failed to turn on fan", zap.Error(err))
	}
	_, err = s.Home(ctx, &proto.HomeRequest{})
	if err != nil {
		logger.Fatal("Failed to home", zap.Error(err))
	}
	_, err = s.FanOff(ctx, &proto.FanOffRequest{})
	if err != nil {
		logger.Fatal("Failed to turn off fan", zap.Error(err))
	}
	pos := float32(100)
	spd := float32(500)
	_, err = s.Move(ctx, &proto.MoveRequest{
		X:     &pos,
		Y:     &pos,
		Z:     &pos,
		Speed: &spd,
		E:     &pos,
	})
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
	<-ctx.Done()
}
