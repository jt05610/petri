package autosampler

import (
	"bytes"
	"context"
	"github.com/jt05610/petri/comm/serial"
	"go.uber.org/zap"
	"io"
	"sync/atomic"
	"time"
)

type State uint32

const (
	Idle State = iota
	Injecting
	Finishing
)

type Settings map[string]Byteable

type Server struct {
	*Parser
	Cts            atomic.Bool
	logger         *zap.Logger
	state          *atomic.Uint32
	port           *serial.Port
	pendingRequest *atomic.Pointer[Request]
	rxChan         <-chan io.Reader
	TxChan         chan []byte
	listenCancel   context.CancelFunc
	settings       *atomic.Pointer[Settings]
}

func (s *Server) currentState() State {
	return State(s.state.Load())
}

func (s *Server) do(cmd []byte) error {
	msg := append(cmd, '\r')
	s.logger.Debug("Sending message", zap.String("msg", string(msg)))
	s.TxChan <- msg
	s.Cts.Store(false)
	for !s.Cts.Load() {
	}
	return nil
}

func New(port *serial.Port, logger *zap.Logger) *Server {
	txCh := make(chan []byte, 100)
	rxCh, err := port.ChannelPort(context.Background(), txCh, '\r', '\n')
	if err != nil {
		panic(err)
	}
	ctx, can := context.WithCancel(context.Background())
	s := &Server{
		port:           port,
		rxChan:         rxCh,
		Cts:            atomic.Bool{},
		logger:         logger,
		state:          new(atomic.Uint32),
		pendingRequest: new(atomic.Pointer[Request]),
		TxChan:         txCh,
		listenCancel:   can,
		settings:       new(atomic.Pointer[Settings]),
	}
	s.Cts.Store(true)
	s.state.Store(uint32(Idle))
	s.settings.Store(&Settings{})
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

func (s *Server) ProcessResponse(resp *Response) {
	if resp.Request == nil {
		return
	}
	switch resp.Request.Header {
	case Inject:
		s.logger.Info("Received inject message")
		s.state.Store(uint32(Finishing))
	case Get:
		if resp.Request.Id == InjectionStatus.Id {
			if v, ok := resp.Data.(IntData); !ok {
				s.logger.Error("Failed to parse injection status")
			} else {
				if v == IntData(0) {
					s.state.Store(uint32(Idle))
				}
			}
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
			buf := bytes.NewBuffer(append(bb, '\r'))
			parser := NewParser(buf)
			rr, err := parser.Parse()
			if err != nil {
				// write back to the buffer so that we can try again
				s.logger.Error("Failed to parse message", zap.Error(err))
				continue
			}
			for _, r := range rr {
				s.ProcessResponse(r)
			}
			s.Cts.Store(true)
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
			for _, r := range HeartbeatMsgs() {
				err := s.do(r.Bytes())
				if err != nil {
					s.logger.Error("Failed to send heartbeat", zap.Error(err))
				}
			}
		}
	}
}

func (s *Server) Send(r *Request) error {
	s.pendingRequest.Store(r)
	return s.do(r.Bytes())
}

func (s *Server) UpdateSettings(ctx context.Context, r *InjectRequest) error {
	newSettings, err := SettingsMap(r)
	if err != nil {
		return err
	}
	s.settings.Store(newSettings)
	return nil
}

func SettingsMap(r *InjectRequest) (*Settings, error) {
	newSettings := make(Settings)
	rr, err := r.Requests()
	if err != nil {
		return nil, err
	}
	for _, r := range rr {
		newSettings[r.Name] = r.Data
	}
	return &newSettings, nil
}

func (s *Server) NecessaryRequests(r *InjectRequest) ([]*Request, error) {
	settings := s.settings.Load()
	req, err := r.Requests()
	if err != nil {
		return nil, err
	}
	toSend := make([]*Request, 0)
	for i, r := range req {
		current, found := (*settings)[r.Name]
		if !found {
			toSend = append(toSend, req[i])
			continue
		}
		if a, ok := current.(IntArrayData); ok {
			b, ok := r.Data.(IntArrayData)
			if !ok {
				toSend = append(toSend, req[i])
				continue
			}
			if len(a) != len(b) {
				toSend = append(toSend, req[i])
				continue
			}
			for j, v := range a {
				if v != b[j] {
					toSend = append(toSend, req[i])
					break
				}
			}
			continue
		}
		if current != r.Data {
			toSend = append(toSend, req[i])
		}
	}
	return toSend, nil
}

func (s *Server) startInjection(ctx context.Context, r *InjectRequest) error {
	toSend, err := s.NecessaryRequests(r)
	if err != nil {
		return err
	}
	for _, r := range toSend {
		err := s.Send(r)
		if err != nil {
			return err
		}
	}
	err = s.UpdateSettings(ctx, r)
	if err != nil {
		return err
	}
	return s.Send(DoInjection)
}
