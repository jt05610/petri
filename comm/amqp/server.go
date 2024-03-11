package amqp

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
)

type Server struct {
	*petri.Net
}

func NewServer(n *petri.Net) *Server {
	return &Server{n}
}

type Handlers map[string]petri.Handler

func (s *Server) RegisterHandlers(h Handlers) {
	s.makeQueues()
}

func (s *Server) makeQueues() {
	for p := range s.Places {
		fmt.Printf("place: %s\n", p)
	}
}

func (s *Server) Serve(ctx context.Context) error {
	return nil
}
