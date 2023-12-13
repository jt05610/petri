package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jt05610/petri/comm/grbl"
	"github.com/jt05610/petri/comm/grbl/proto/v1"
	"github.com/jt05610/petri/comm/serial"
	"go.uber.org/zap"
	"io"
	"math"
	"sync/atomic"
	"time"
)

var _ v1.GRBLServer = (*Server)(nil)

type Server struct {
	*grbl.Parser
	Cts           atomic.Bool
	logger        *zap.Logger
	machineStatus *atomic.Pointer[grbl.Status]
	state         *atomic.Pointer[v1.State]
	port          *serial.Port
	rxChan        <-chan io.Reader
	TxChan        chan []byte
	listenCancel  context.CancelFunc
	v1.UnimplementedGRBLServer
}

func (s *Server) currentState() *v1.State {
	return s.state.Load()
}

func (s *Server) status() *grbl.Status {
	return s.machineStatus.Load()
}

func (s *Server) do(cmd []byte, synchronous bool, check func(state *v1.State) bool) error {
	for !s.Cts.Load() {
	}
	s.logger.Debug("Sending command", zap.String("cmd", string(cmd)))
	s.TxChan <- cmd
	if synchronous {
		for {
			status := s.currentState()
			if s.Cts.Load() && check(status) {
				return nil
			}
			if status == nil {
				continue
			}

			if status.Error != nil {
				if status.Error.Error != nil {
					if *status.Error.Error != v1.ErrorCode_ErrorCode_NoError {
						return fmt.Errorf("error: %d", status.Error.Message)
					}
				}
			}
			if status.Alarm != nil {
				if status.Alarm.Alarm != nil {
					if *status.Alarm.Alarm != v1.AlarmCode_AlarmCode_NoAlarm {
						return fmt.Errorf("alarm: %d", status.Alarm.Message)
					}
				}
			}

		}
	}
	return nil
}

func (s *Server) Home(ctx context.Context, req *v1.HomeRequest) (*v1.Response, error) {
	if req.Axis == nil {
		s.TxChan <- []byte("$H\n")
	} else {
		if v, ok := req.Axis.(*v1.HomeRequest_X); ok {
			if v.X {
				s.TxChan <- []byte("$HX\n")
			}
		} else if v, ok := req.Axis.(*v1.HomeRequest_Y); ok {
			if v.Y {
				s.TxChan <- []byte("$HY\n")
			}
		} else if v, ok := req.Axis.(*v1.HomeRequest_Z); ok {
			if v.Z {
				s.TxChan <- []byte("$HZ\n")
			}
		}
	}
	// wait for state to be homing, then idle again
	homingSeen := false
	idleSeen := false
	for !homingSeen && !idleSeen {
		time.Sleep(100 * time.Millisecond)
		status := s.status()
		if status == nil {
			continue
		}
		if status.State == "home" {
			homingSeen = true
			if status.State == "idle" {
				idleSeen = true
			}
		}
	}
	return &v1.Response{}, nil
}

const waitFor = " unlock]"

func New(port *serial.Port, logger *zap.Logger) *Server {
	buf := new(bytes.Buffer)
	txCh := make(chan []byte, 100)
	rxCh, err := port.ChannelPort(context.Background(), txCh)
	if err != nil {
		panic(err)
	}
	ctx, can := context.WithCancel(context.Background())
	for waiting := true; waiting; {
		select {
		case <-ctx.Done():
			panic("context cancelled")
		case msgRdr := <-rxCh:
			msg, err := io.ReadAll(msgRdr)
			if err != nil {
				panic(err)
			}
			logger.Debug("Received message", zap.String("msg", string(msg)))
			if bytes.Contains(msg, []byte(waitFor)) {
				txCh <- []byte("$X\n")
			}
			if bytes.Contains(msg, []byte("[MSG:Caution: Unlocked]")) {
				waiting = false
			}
		}
	}

	s := &Server{
		Parser:        grbl.NewParser(buf),
		port:          port,
		rxChan:        rxCh,
		Cts:           atomic.Bool{},
		logger:        logger,
		state:         &atomic.Pointer[v1.State]{},
		machineStatus: &atomic.Pointer[grbl.Status]{},
		TxChan:        txCh,
		listenCancel:  can,
	}
	s.Cts.Store(false)
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

func (s *Server) StateStream(req *v1.StateStreamRequest, server v1.GRBL_StateStreamServer) error {
	for {
		select {
		case <-server.Context().Done():
			return nil
		case <-time.After(1 * time.Second):
			state := s.currentState()
			if state == nil {
				continue
			}
			err := server.Send(&v1.StateStreamResponse{
				State:     state,
				Timestamp: time.Now().Format(time.RFC3339Nano),
			})
			if err != nil {
				return err
			}
		}
	}
}

func goToMsg(pos *v1.MoveRequest) []byte {
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
	bld.WriteString("\n")
	ret := bld.Bytes()
	fmt.Println("writing: ", string(ret))
	return ret
}

const threshold = 0.02

func floatEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) <= threshold
}

func (s *Server) Move(ctx context.Context, req *v1.MoveRequest) (*v1.Response, error) {
	err := s.do(goToMsg(req), true, func(state *v1.State) bool {
		if req.X != nil {
			if floatEqual(state.Position.X, *req.X) && s.machineStatus.Load().State == "idle" {
				return true
			}
		}
		if req.Y != nil {
			if floatEqual(state.Position.Y, *req.Y) && s.machineStatus.Load().State == "idle" {
				return true
			}
		}
		if req.Z != nil {
			if floatEqual(state.Position.Z, *req.Z) && s.machineStatus.Load().State == "idle" {
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return &v1.Response{
		Message: "ok",
		Response: &v1.Response_Move{
			Move: &v1.MoveResponse{
				Message: "ok",
			},
		},
	}, nil
}

func (s *Server) SpindleOn(ctx context.Context, req *v1.SpindleOnRequest) (*v1.Response, error) {
	err := s.do([]byte("M3\n"), true, func(state *v1.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &v1.Response{
		Message:  "ok",
		Response: &v1.Response_SpindleOn{SpindleOn: &v1.SpindleOnResponse{Message: "ok"}},
	}, nil
}

func (s *Server) SpindleOff(ctx context.Context, req *v1.SpindleOffRequest) (*v1.Response, error) {
	err := s.do([]byte("M5\n"), true, func(state *v1.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &v1.Response{
		Message:  "ok",
		Response: &v1.Response_SpindleOff{SpindleOff: &v1.SpindleOffResponse{Message: "ok"}},
	}, nil
}

func (s *Server) MistOn(ctx context.Context, req *v1.MistOnRequest) (*v1.Response, error) {
	err := s.do([]byte("M7\n"), true, func(state *v1.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &v1.Response{
		Message: "ok",
		Response: &v1.Response_MistOn{
			MistOn: &v1.MistOnResponse{
				Message: "ok",
			},
		},
	}, nil
}

func (s *Server) FloodOn(ctx context.Context, req *v1.FloodOnRequest) (*v1.Response, error) {
	err := s.do([]byte("M8\n"), true, func(state *v1.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &v1.Response{
		Message: "ok",
		Response: &v1.Response_FloodOn{
			FloodOn: &v1.FloodOnResponse{
				Message: "ok",
			},
		},
	}, nil
}

func (s *Server) CoolantOff(ctx context.Context, req *v1.CoolantOffRequest) (*v1.Response, error) {
	err := s.do([]byte("M9\n"), true, func(state *v1.State) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	return &v1.Response{
		Message: "ok",
		Response: &v1.Response_CoolantOff{
			CoolantOff: &v1.CoolantOffResponse{
				Message: "ok",
			},
		},
	}, nil
}

var (
	HeartbeatMsg = "?\n"
)

func (s *Server) UpdateStatus(status grbl.StatusUpdate) {
	if _, ok := status.(*grbl.Ack); ok {
		s.Cts.Store(true)
		return
	}
	if grblErr, ok := status.(grbl.Error); ok {
		s.logger.Fatal("Received error", zap.Int("msg", int(grblErr)))
		return
	}
	if grblAlarm, ok := status.(grbl.Alarm); ok {
		s.logger.Fatal("Received alarm", zap.Int("msg", int(grblAlarm)))
		return
	}
	if upd, ok := status.(*grbl.Status); ok {
		s.machineStatus.Store(upd)
		newState := &v1.State{
			Position: &v1.Position{
				X: upd.MachinePosition.X,
				Y: upd.MachinePosition.Y,
				Z: upd.MachinePosition.Z,
			},
		}
		if upd.Error != nil {
			ec := v1.ErrorCode(*upd.Error)
			s := ec.String()
			newState.Error = &v1.Error{
				Message: &s,
				Error:   &ec,
			}
		}
		if upd.Override != nil {
			fOvr := uint32(upd.Override.Feed)
			rOvr := uint32(upd.Override.Rapid)
			sOvr := uint32(upd.Override.Spindle)
			newState.Offsets = &v1.Offsets{
				Feed:    &fOvr,
				Rapid:   &rOvr,
				Spindle: &sOvr,
			}
		}
		if upd.Active != nil {
			newState.Active = make([]v1.Peripheral, 0, 4)
			if upd.Active.Flood {
				newState.Active = append(newState.Active, v1.Peripheral_Flood)
			}
			if upd.Active.Mist {
				newState.Active = append(newState.Active, v1.Peripheral_Mist)
			}
			if upd.Active.Spindle {
				newState.Active = append(newState.Active, v1.Peripheral_Spindle)
			}
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
			s.logger.Debug("Received response", zap.String("response", string(bb)))
			buf := bytes.NewBuffer(bb)
			parser := grbl.NewParser(buf)
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

func (s *Server) RunHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	errCount := 3
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			errCount = 3
			if !s.Cts.Load() {
				errCount--
				if errCount == 0 {
					s.logger.Fatal("Failed to receive ack")
				}
				continue
			}
			s.Cts.Store(false)
			s.TxChan <- []byte(HeartbeatMsg)
		}
	}
}
