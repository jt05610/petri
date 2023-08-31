package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/comm/serial"
	"github.com/jt05610/petri/grbl"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"go.uber.org/zap"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var _ proto.GRBLServer = (*Server)(nil)

type Server struct {
	*grbl.Parser
	cts           atomic.Bool
	logger        *zap.Logger
	machineStatus *atomic.Pointer[grbl.Status]
	state         *atomic.Pointer[proto.State]
	port          *serial.Port
	rxChan        <-chan []byte
	txChan        chan []byte
	rxBuf         *bytes.Buffer
	listenCancel  context.CancelFunc
	proto.UnimplementedGRBLServer
}

func (s *Server) currentState() *proto.State {
	return s.state.Load()
}

func (s *Server) status() *grbl.Status {
	return s.machineStatus.Load()
}

func (s *Server) do(cmd []byte, synchronous bool, check func(state *proto.State) bool) error {
	for !s.cts.Load() {
	}
	s.txChan <- cmd
	if synchronous {
		for {
			time.Sleep(100 * time.Millisecond)
			status := s.currentState()
			if status == nil {
				continue
			}

			if status.Error != nil {
				if status.Error.Error != nil {
					if *status.Error.Error != proto.ErrorCode_ErrorCode_NoError {
						return fmt.Errorf("error: %d", status.Error.Message)
					}
				}
			}
			if status.Alarm != nil {
				if status.Alarm.Alarm != nil {
					if *status.Alarm.Alarm != proto.AlarmCode_AlarmCode_NoAlarm {
						return fmt.Errorf("alarm: %d", status.Alarm.Message)
					}
				}
			}
			if check(status) {
				return nil
			}
		}
	}
	return nil
}

func (s *Server) Home(ctx context.Context, req *proto.HomeRequest) (*proto.Response, error) {
	s.txChan <- []byte("$H\n")
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
	return &proto.Response{}, nil
}

const waitFor = " unlock]\r\n"

func New(port *serial.Port, logger *zap.Logger) *Server {
	buf := new(bytes.Buffer)
	txCh := make(chan []byte, 100)
	rxCh, err := port.ChannelPort(context.Background(), txCh)
	if err != nil {
		panic(err)
	}
	ctx, can := context.WithCancel(context.Background())
	seenBuf := new(bytes.Buffer)
	for waiting := true; waiting; {
		select {
		case <-ctx.Done():
			panic("context cancelled")
		case msg := <-rxCh:
			seenBuf.Write(msg)
			if bytes.Contains(msg, []byte("\n")) {
				logger.Debug("Received message", zap.String("msg", string(msg)))
				seenBytes := seenBuf.Bytes()
				if bytes.Contains(seenBytes, []byte(waitFor)) {
					txCh <- []byte("$X\n")
				}
				if bytes.Contains(seenBytes, []byte("[MSG:Caution: Unlocked]\r")) {
					waiting = false
				} else {
					seenBuf.Reset()
					seenBuf.Write(seenBytes)
				}
			}
		}
	}

	s := &Server{
		Parser:        grbl.NewParser(buf),
		port:          port,
		rxBuf:         new(bytes.Buffer),
		rxChan:        rxCh,
		cts:           atomic.Bool{},
		logger:        logger,
		state:         &atomic.Pointer[proto.State]{},
		machineStatus: &atomic.Pointer[grbl.Status]{},
		txChan:        txCh,
		listenCancel:  can,
	}
	s.cts.Store(false)
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

func (s *Server) StateStream(req *proto.StateStreamRequest, server proto.GRBL_StateStreamServer) error {
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
	bld.WriteString("\n")
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

func (s *Server) SpindleOn(ctx context.Context, req *proto.SpindleOnRequest) (*proto.Response, error) {
	err := s.do([]byte("M3\n"), true, func(state *proto.State) bool {
		for _, p := range state.Active {
			if p == proto.Peripheral_Spindle {
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message:  "ok",
		Response: &proto.Response_SpindleOn{SpindleOn: &proto.SpindleOnResponse{Message: "ok"}},
	}, nil
}

func (s *Server) SpindleOff(ctx context.Context, req *proto.SpindleOffRequest) (*proto.Response, error) {
	err := s.do([]byte("M5\n"), true, func(state *proto.State) bool {
		if state.Active == nil || len(state.Active) == 0 {
			return true
		}
		for _, p := range state.Active {
			if p == proto.Peripheral_Spindle {
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message:  "ok",
		Response: &proto.Response_SpindleOff{SpindleOff: &proto.SpindleOffResponse{Message: "ok"}},
	}, nil
}

func (s *Server) MistOn(ctx context.Context, req *proto.MistOnRequest) (*proto.Response, error) {
	err := s.do([]byte("M7\n"), true, func(state *proto.State) bool {
		for _, p := range state.Active {
			if p == proto.Peripheral_Mist {
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message: "ok",
		Response: &proto.Response_MistOn{
			MistOn: &proto.MistOnResponse{
				Message: "ok",
			},
		},
	}, nil
}

func (s *Server) FloodOn(ctx context.Context, req *proto.FloodOnRequest) (*proto.Response, error) {
	err := s.do([]byte("M8\n"), true, func(state *proto.State) bool {
		for _, p := range state.Active {
			if p == proto.Peripheral_Flood {
				return true
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return &proto.Response{
		Message: "ok",
		Response: &proto.Response_FloodOn{
			FloodOn: &proto.FloodOnResponse{
				Message: "ok",
			},
		},
	}, nil
}

func (s *Server) CoolantOff(ctx context.Context, req *proto.CoolantOffRequest) (*proto.Response, error) {
	err := s.do([]byte("M9\n"), true, func(state *proto.State) bool {
		for _, p := range state.Active {
			if state.Active == nil || len(state.Active) == 0 {
				return true
			}
			if p == proto.Peripheral_Flood || p == proto.Peripheral_Mist {
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
		Response: &proto.Response_CoolantOff{
			CoolantOff: &proto.CoolantOffResponse{
				Message: "ok",
			},
		},
	}, nil
}

var (
	HeartbeatMsg = "?\n"
)

func (s *Server) UpdateStatus(status grbl.StatusUpdate) {
	if ack, ok := status.(*grbl.Ack); ok {
		s.cts.Store(true)
		s.logger.Debug("Received ack", zap.String("msg", ack.String()))
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
		newState := &proto.State{
			Position: &proto.Position{
				X: upd.MachinePosition.X,
				Y: upd.MachinePosition.Y,
				Z: upd.MachinePosition.Z,
			},
		}
		if upd.Error != nil {
			ec := proto.ErrorCode(*upd.Error)
			s := ec.String()
			newState.Error = &proto.Error{
				Message: &s,
				Error:   &ec,
			}
		}
		if upd.Override != nil {
			fOvr := uint32(upd.Override.Feed)
			rOvr := uint32(upd.Override.Rapid)
			sOvr := uint32(upd.Override.Spindle)
			newState.Offsets = &proto.Offsets{
				Feed:    &fOvr,
				Rapid:   &rOvr,
				Spindle: &sOvr,
			}
		}
		if upd.Active != nil {
			newState.Active = make([]proto.Peripheral, 0, 4)
			if upd.Active.Flood {
				newState.Active = append(newState.Active, proto.Peripheral_Flood)
			}
			if upd.Active.Mist {
				newState.Active = append(newState.Active, proto.Peripheral_Mist)
			}
			if upd.Active.Spindle {
				newState.Active = append(newState.Active, proto.Peripheral_Spindle)
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
			s.rxBuf.Write(msg)
			// if we have a newline, parse the message
			if bytes.Contains(msg, []byte("\n")) {
				// split the message on newlines
				msgs := bytes.Split(s.rxBuf.Bytes(), []byte("\n"))
				s.rxBuf.Reset()
				if msg[len(msg)-1] != '\n' {
					// if the last message is not a newline, we have a partial message so we need to keep it in the buffer
					s.rxBuf.Write(msgs[len(msgs)-1])
					msgs = msgs[:len(msgs)-1]
				}
				for _, msg := range msgs {
					s.logger.Debug("Received message", zap.String("msg", string(msg)))
					buf := bytes.NewReader(msg)
					parser := grbl.NewParser(buf)
					upd, err := parser.Parse()

					if err != nil {
						// write back to the buffer so that we can try again
						s.logger.Error("Failed to parse message", zap.Error(err), zap.String("msg", string(msg)))
					} else {

						s.UpdateStatus(upd)
					}
				}
			}
		}
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

func getVal(name string, s *string) error {
	v, found := os.LookupEnv(name)
	if !found {
		return fmt.Errorf("environment variable %s not found", name)
	}
	*s = v
	return nil
}

type Environment struct {
	SerialPort   string
	Baud         int
	StartupBlock string
}

func load() *Environment {
	vals := map[string]*string{
		"SERIAL_PORT":   new(string),
		"SERIAL_BAUD":   new(string),
		"STARTUP_BLOCK": new(string),
	}
	for k, v := range vals {
		err := getVal(k, v)
		if err != nil {
			panic(err)
		}
	}
	baud, err := strconv.ParseInt(*vals["SERIAL_BAUD"], 10, 64)
	if err != nil {
		panic(err)
	}
	return &Environment{
		SerialPort:   *vals["SERIAL_PORT"],
		Baud:         int(baud),
		StartupBlock: *vals["STARTUP_BLOCK"],
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
	_, err = s.Home(ctx, &proto.HomeRequest{})
	s.txChan <- []byte("G55\n")
	if err != nil {
		logger.Fatal("Failed to home", zap.Error(err))
	}
	for !s.cts.Load() {

	}
	zVal := float32(-12)
	feed := float32(500)
	_, err = s.Move(ctx, &proto.MoveRequest{
		Z:     &zVal,
		Speed: &feed,
	})
	if err != nil {
		logger.Fatal("Failed to move", zap.Error(err))
	}
	resp, err := s.SpindleOn(ctx, &proto.SpindleOnRequest{})
	if err != nil {
		logger.Fatal("Failed to turn on spindle", zap.Error(err))
	}
	logger.Info("Spindle on", zap.Any("resp", resp))
	resp, err = s.MistOn(ctx, &proto.MistOnRequest{})
	if err != nil {
		logger.Fatal("Failed to turn on mist", zap.Error(err))
	}
	logger.Info("Mist on", zap.Any("resp", resp))
	resp, err = s.CoolantOff(ctx, &proto.CoolantOffRequest{})
	if err != nil {
		logger.Fatal("Failed to turn off coolant", zap.Error(err))
	}
	logger.Info("Coolant off", zap.Any("resp", resp))
	resp, err = s.SpindleOff(ctx, &proto.SpindleOffRequest{})
	if err != nil {
		logger.Fatal("Failed to turn off spindle", zap.Error(err))
	}
	logger.Info("spindle off", zap.Any("resp", resp))
	resp, err = s.FloodOn(ctx, &proto.FloodOnRequest{})
	if err != nil {
		logger.Fatal("Failed to turn on flood", zap.Error(err))
	}

	logger.Info("Flood on", zap.Any("resp", resp))
	resp, err = s.CoolantOff(ctx, &proto.CoolantOffRequest{})
	if err != nil {
		logger.Fatal("Failed to turn off coolant", zap.Error(err))
	}
	logger.Info("Coolant off", zap.Any("resp", resp))

	newTarget := float32(0)
	newSpeed := float32(500)
	_, err = s.Move(ctx, &proto.MoveRequest{
		Z:     &newTarget,
		Speed: &newSpeed,
	})

	if err != nil {
		logger.Fatal("Failed to move", zap.Error(err))
	}
	logger.Info("Moved", zap.Any("resp", resp))
	<-ctx.Done()
}
