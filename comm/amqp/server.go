package amqp

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp/queue"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type Server struct {
	*petri.Net
	*amqp.Channel
	Queues  map[string]*queue.Local
	Initial petri.Marking
	mu      sync.Mutex
}

func (s *Server) makeQueue(name string) {
	q, err := s.Channel.QueueDeclare(
		name,
		true,
		false,

		false,
		true,
		nil,
	)
	if err != nil {
		panic(err)
	}
	err = s.Channel.QueueBind(
		"",
		name,
		s.Name,
		false,
		nil,
	)
	if err != nil {
		panic(err)

	}
	s.Queues[name] = &queue.Local{
		Queue:    q,
		exchange: s.Name,
		ch:       s.Channel,
		Schema:   s.InputSchema(name),
	}
}

func NewServer(conn *amqp.Connection, n *petri.Net) *Server {

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	err = ch.ExchangeDeclare(n.Name, "topic", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	srv := &Server{
		Net:     n,
		Channel: ch,
		Queues:  make(map[string]*queue.Local),
	}
	for tn, t := range n.Transitions {
		if t.Cold {
			srv.makeQueue(tn)
		}
	}
	srv.makeQueues()
	return srv
}

type Handlers map[string]petri.Handler

func (s *Server) RegisterHandlers(h Handlers) {
	for name, handler := range h {
		t := s.Transition(name)
		if t == nil {
			panic(fmt.Errorf("no transition named %s", name))
		}
		t.Handler = handler
	}
}

func (s *Server) makeQueues() *Server {
	for _, place := range s.InPlaces() {
		fmt.Printf("Making queue for place: %s\n", place.ID)
		q, err := s.Channel.QueueDeclare(
			place.ID,
			true,
			false,
			false,
			true,
			nil,
		)
		if err != nil {
			panic(err)
		}
		s.Queues[place.ID] = &queue.Local{
			Queue:    q,
			exchange: s.Name,
			ch:       s.Channel,
			Schema:   place.AcceptedTokens[0],
		}
		fmt.Printf("place: %s\n", place.ID)
	}
	return s
}

func (s *Server) UpdateMarking(m petri.Marking, pl string, t *petri.Token) (petri.Marking, error) {
	place := s.Place(pl)
	if place != nil {
		err := m.PlaceTokens(s.Place(pl), t)
		if err != nil {
			return nil, err
		}
		return m, nil
	}
	return s.Process(m, petri.Event[any]{
		Name: pl,
		Data: t.Bytes(),
	})
}

var _ error = (*MultiError)(nil)

type MultiError []error

func (m MultiError) Error() string {
	var s string
	for _, e := range m {
		s += e.Error() + "\n"
	}
	return s
}

func (s *Server) Serve(ctx context.Context) error {
	var wg sync.WaitGroup
	errs := make(chan error, 100)
	marking := s.Initial
	for qn, q := range s.Queues {
		fmt.Printf("listening on queue: %s\n", qn)
		wg.Add(1)
		go func(qn string, q *queue.Local) {
			defer wg.Done()
			messages, err := q.ch.Consume(qn, "", true, false, false, false, nil)
			if err != nil {
				errs <- err
			}
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-messages:
					fmt.Printf("Received message on %s\n", qn)
					body := msg.Body
					if body == nil {
						body = []byte{}
					}
					t, err := q.Schema.NewToken(body)
					if err != nil {
						errs <- err
						return
					}
					s.mu.Lock()
					marking, err = s.UpdateMarking(marking, qn, t)
					if err != nil {
						errs <- err
						return
					}
					fmt.Println(marking)
					s.mu.Unlock()
				}
			}
		}(qn, q)
	}
	wg.Wait()
	close(errs)
	ret := make(MultiError, 0, 100)
	for e := range errs {
		ret = append(ret, e)
	}
	if len(ret) > 0 {
		return ret
	}
	return nil
}
