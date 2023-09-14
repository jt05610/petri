package marlin

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jt05610/petri/comm/serial"
	proto "github.com/jt05610/petri/marlin/proto/v1"
	"go.uber.org/zap"
	"io"
	"sync/atomic"
	"time"
)

var _ proto.MarlinServer = (*Server)(nil)

type Server struct {
	*Parser
	cts           atomic.Bool
	logger        *zap.Logger
	machineStatus *atomic.Pointer[Status]
	state         *atomic.Pointer[proto.State]
	port          *serial.Port
	rxChan        <-chan io.Reader
	TxChan        chan []byte
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

func (s *Server) status() *Status {
	return s.machineStatus.Load()
}

func (s *Server) do(cmd []byte, synchronous bool, check func(state *proto.State) bool) error {
	for !s.cts.Load() {
	}
	s.TxChan <- cmd
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
		s.TxChan <- []byte("G28\n")
	} else {
		if v, ok := req.Axis.(*proto.HomeRequest_X); ok {
			if v.X {
				s.TxChan <- []byte("G28 X\n")
			}
		} else if v, ok := req.Axis.(*proto.HomeRequest_Y); ok {
			if v.Y {
				s.TxChan <- []byte("G28 Y\n")
			}
		} else if v, ok := req.Axis.(*proto.HomeRequest_Z); ok {
			if v.Z {
				s.TxChan <- []byte("G28 Z\n")
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

func New(ctx context.Context, port *serial.Port, logger *zap.Logger) *Server {
	buf := new(bytes.Buffer)
	txCh := make(chan []byte, 100)
	rxCh, err := port.ChannelPort(ctx, txCh)
	if err != nil {
		panic(err)
	}

	s := &Server{
		Parser:        NewParser(buf),
		port:          port,
		rxChan:        rxCh,
		cts:           atomic.Bool{},
		logger:        logger,
		state:         &atomic.Pointer[proto.State]{},
		machineStatus: &atomic.Pointer[Status]{},
		TxChan:        txCh,
	}

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
	HeartbeatMsg = "M114\n"
)

func (s *Server) UpdateStatus(status StatusUpdate) {
	if _, ok := status.(*Ack); ok {
		s.cts.Store(true)
		return
	}
	if _, ok := status.(*Processing); ok {
		return
	}

	if upd, ok := status.(*Status); ok {
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

func (s *Server) RunHeartbeat(ctx context.Context) {
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
			s.logger.Debug("Sending heartbeat")
			s.TxChan <- []byte(HeartbeatMsg)
		}
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
			parser := NewParser(buf)
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
