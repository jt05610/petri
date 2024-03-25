package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp/queue"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"sync"
	"time"
)

type Listener interface {
	Listen(ctx context.Context) error
	Close() error
}

type Queue struct {
	petri.TokenQueue
	PlaceID  string
	ctx      context.Context
	cancel   func()
	ch       chan<- struct{}
	deferral func()
}

type MarkingUpdate struct {
	PlaceID string
	Tokens  []petri.Token
}

func NewQueue(placeID string, q petri.TokenQueue, deferral func(), signal chan<- struct{}) *Queue {
	return &Queue{
		TokenQueue: q,
		ch:         signal,
		PlaceID:    placeID,
		deferral:   deferral,
	}
}

func (q *Queue) Close() {
	q.cancel()
	q.deferral()
}

type RPC struct {
	amqp.Delivery
}

type Bindings map[string]HandlerFunc

type Router struct {
	petri.Marking
	mu sync.Mutex
	Bindings
	*petri.Net
	*amqp.Channel
	Exchange string
	Name     string
	Timeout  time.Duration
	ch       chan struct{}
	ctx      context.Context
	cancel   func()
}

func (r *Router) Handle(ctx context.Context, m amqp.Delivery) {
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	if h, ok := r.Bindings[m.RoutingKey]; ok {
		h(ctx, Delivery{
			Delivery: m,
			Channel:  r.Channel,
		})
	}
}

func (r *Router) Bind() error {
	for route := range r.Bindings {
		err := r.QueueBind(r.Name, route, r.Exchange, false, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewRouter(ch *amqp.Channel, initial petri.Marking, name, exchange string) (*Router, error) {
	r := &Router{
		Channel:  ch,
		Exchange: exchange,
		Name:     name,
		Marking:  initial,
	}
	r.Bindings = r.MakeBindings()
	return r, nil
}

func (r *Router) Mark(marking petri.Marking) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Marking = marking
}

func (r *Router) Listen(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)
	r.ch = make(chan struct{})
	var err error
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.ch:
			r.mu.Lock()
			r.Marking, err = r.Net.Process(r.Marking)
			if err != nil {
				if !errors.Is(petri.ErrNoEvents, err) {
					slog.Error("error processing marking", slog.Any("message", err))
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *Router) Route(route string, h HandlerFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Bindings[route] = h
	return r.QueueBind(r.Name, route, r.Exchange, false, nil)
}

func (r *Router) UnRoute(route string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.Bindings[route]; !ok {
		return r.QueueUnbind(r.Name, route, r.Exchange, nil)
	}
	delete(r.Bindings, route)
	return r.QueueUnbind(r.Name, route, r.Exchange, nil)
}

type HandlerFunc func(context.Context, Delivery)

type Codec[T any] interface {
	Marshal(t *T) ([]byte, error)
	Unmarshal([]byte, *T) error
}

type JSONCodec[T any] struct{}

func (j JSONCodec[T]) Marshal(t *T) ([]byte, error) {
	return json.Marshal(t)
}

func (j JSONCodec[T]) Unmarshal(b []byte, t *T) error {
	return json.Unmarshal(b, t)
}

type RPCCodec[T, U any] struct {
	Request  Codec[T]
	Response Codec[U]
}

type Delivery struct {
	amqp.Delivery
	*amqp.Channel
}

func NewHandler[T, U any](exchange string, codec RPCCodec[T, U], f func(ctx context.Context, t *T) (*U, error)) HandlerFunc {
	return func(ctx context.Context, m Delivery) {
		defer func() {
			err := m.Delivery.Ack(false)
			if err != nil {
				slog.Error("error acknowledging message", slog.Any("message", err))
			}
		}()
		var t T
		err := codec.Request.Unmarshal(m.Body, &t)
		if err != nil {
			slog.Error("error unmarshalling message", slog.Any("message", err))
			return
		}
		u, err := f(ctx, &t)
		if err != nil {
			slog.Error("error handling message", slog.Any("message", err))
			return
		}
		if m.ReplyTo == "" {
			return
		}
		b, err := codec.Response.Marshal(u)
		if err != nil {
			slog.Error("error marshalling response", slog.Any("message", err))
			return
		}
		err = m.PublishWithContext(ctx, exchange, m.ReplyTo, false, false, amqp.Publishing{
			ContentType:   "application/json",
			Body:          b,
			Timestamp:     time.Now(),
			CorrelationId: m.CorrelationId,
		})
		if err != nil {
			slog.Error("error publishing response", slog.Any("message", err))
		}
	}
}

type ArcDirection int

const (
	ToPlace ArcDirection = iota
	FromPlace
	ToAndFromPlace
)

// ConnectRequest is a request to connect a transition to a place on another petri net.
type ConnectRequest struct {
	Url            string
	Exchange       string
	PlaceName      string
	TransitionName string
	Expression     string
	Direction      ArcDirection
}

type Response struct {
	Ok    bool
	Error string
}

var ConnectCodec = RPCCodec[ConnectRequest, Response]{
	Request:  JSONCodec[ConnectRequest]{},
	Response: JSONCodec[Response]{},
}
var DisconnectCodec = RPCCodec[DisconnectRequest, Response]{
	Request:  JSONCodec[DisconnectRequest]{},
	Response: JSONCodec[Response]{},
}

// DisconnectRequest is a request to disconnect a transition from a place on another petri net.
type DisconnectRequest struct {
	Url       string
	Exchange  string
	PlaceName string
}

// handleConnect connects a transition to a place on another petri net using an amqp.Remote queue. It begins monitoring the queue for changes in the marking of the place.
func (r *Router) handleConnect(_ context.Context, request *ConnectRequest) (*Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tr := r.Transition(request.TransitionName)
	if tr == nil {
		transitions := make([]string, 0, len(r.Transitions))
		for _, t := range r.Transitions {
			transitions = append(transitions, t.ID)
		}
		return &Response{
			Ok:    false,
			Error: fmt.Sprintf("no transition named %s. Options are %v", request.TransitionName, transitions),
		}, nil
	}
	// need to make a new place for supplying the transition's desired event schema and arc for the petri net
	pl := petri.NewPlace(request.PlaceName, 1, tr.EventSchema)
	pl.ID = request.PlaceName
	r.Net = r.Net.WithPlaces(pl)
	switch request.Direction {
	case ToPlace:
		r.Net = r.WithArcs(petri.NewArc(tr, pl, request.Expression, tr.EventSchema))
	case FromPlace:
		r.Net = r.WithArcs(petri.NewArc(pl, tr, request.Expression, tr.EventSchema))
	case ToAndFromPlace:
		r.Net = r.WithArcs(
			petri.NewArc(pl, tr, request.Expression, tr.EventSchema),
			petri.NewArc(tr, pl, request.Expression, tr.EventSchema),
		)
	}

	// now we need to connect to the remote place on the other petri net
	conn, err := amqp.Dial(request.Url)
	if err != nil {
		return &Response{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	ch, err := conn.Channel()
	if err != nil {
		return &Response{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	q := queue.NewRemote(request.Exchange, ch, pl)
	if current, ok := r.Marking[pl.ID]; ok {
		current.Close()
	}
	d := func() {
		err := conn.Close()
		if err != nil {
			slog.Error("error closing amqp connection", slog.Any("message", err))
		}
		err = ch.Close()
		if err != nil {
			slog.Error("error closing amqp channel", slog.Any("message", err))
		}
	}
	rem := NewQueue(pl.ID, q, d, r.ch)
	r.Marking[pl.ID] = rem
	go func() {
		err := rem.Listen(r.ctx)
		if err != nil {
			slog.Error("error listening", slog.Any("message", err))
		}
	}()
	return &Response{
		Ok: true,
	}, nil
}

// handleDisconnect disconnects a transition from a place on another petri net. It stops monitoring the queue for
// changes in the marking of the place, and closes any connections that were made in connecting to the place.
func (r *Router) handleDisconnect(_ context.Context, request *DisconnectRequest) (*Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pl := r.Place(request.PlaceName)
	if pl == nil {
		return &Response{
			Ok:    false,
			Error: "no place found",
		}, nil
	}

	q := r.Marking[request.PlaceName]
	if q == nil {
		return &Response{
			Ok:    false,
			Error: "no queue found",
		}, nil
	}
	q.Close()
	delete(r.Marking, request.PlaceName)
	for _, input := range r.Inputs(pl) {
		r.Net = r.Net.WithoutArc(input)
	}
	for _, output := range r.Outputs(pl) {
		r.Net = r.Net.WithoutArc(output)
	}
	r.Net = r.Net.WithoutPlace(pl)
	delete(r.Marking, pl.ID)
	return &Response{
		Ok: true,
	}, nil
}

func (r *Router) MakeBindings() Bindings {
	return Bindings{
		"connect":    NewHandler(r.Net.Name, ConnectCodec, r.handleConnect),
		"disconnect": NewHandler(r.Net.Name, DisconnectCodec, r.handleDisconnect),
	}
}

func (q *Queue) Listen(ctx context.Context) error {
	q.ctx, q.cancel = context.WithCancel(ctx)
	go func() {
		defer q.TokenQueue.Close()
		ch := q.Monitor(q.ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				q.ch <- struct{}{}
			}
		}
	}()
	return nil
}

type Server struct {
	*petri.Net
	conn *amqp.Connection
	*Router
	initial  petri.Marking
	bindings map[string]func(context.Context, amqp.Delivery)
	mu       sync.Mutex
	Timeout  time.Duration
	ch       *amqp.Channel
}

func (s *Server) Close() {
	err := s.ch.Close()
	if err != nil {
		slog.Error("error closing amqp channel", slog.Any("message", err))
	}
}

func (s *Server) createEventQueue(t *petri.Transition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	schema := t.EventSchema
	pl := petri.NewPlace(t.Name, 1, schema)
	pl.ID = t.ID
	pl.IsEvent = true
	arc := petri.NewArc(pl, t, schema.Name, schema)
	s.Net = s.Net.WithPlaces(pl).WithArcs(arc)
	q := queue.NewLocal(s.Net.Name, s.ch, pl)
	s.Router.Marking[pl.ID] = q
}

func (s *Server) makeEventQueues() {
	for tn, t := range s.Transitions {
		if !t.Cold || !s.Owns(t) {
			continue
		}
		slog.Info("creating queue", slog.String("name", tn), slog.String("type", "event"))
		s.createEventQueue(t)
	}
}

func NewServer(conn *amqp.Connection, n *petri.Net, initial petri.Marking) *Server {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	err = ch.ExchangeDeclare(n.Name, "topic", false, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		panic(err)
	}
	srv := &Server{
		Net:     n,
		conn:    conn,
		initial: initial,
		ch:      ch,
		Timeout: 30 * time.Second,
	}
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

var _ error = (*MultiError)(nil)

type MultiError []error

func (m MultiError) Error() string {
	var s string
	for _, e := range m {
		s += e.Error() + "\n"
	}
	return s
}

func (s *Server) WithTimeout(t time.Duration) *Server {
	s.Timeout = t
	return s
}

func (s *Server) Serve(ctx context.Context) error {
	var err error
	defer func() {
		err := s.ch.Close()
		if err != nil {
			slog.Error("error closing amqp channel", slog.Any("message", err))
		}
	}()
	q, err := s.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return err
	}
	s.Router, err = NewRouter(s.ch, s.initial, q.Name, s.Net.Name)
	if err != nil {
		return err
	}
	s.Router.Net = s.Net
	s.Router.Timeout = s.Timeout
	err = s.Router.Bind()
	if err != nil {
		return err
	}
	messages, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	go func() {
		err := s.Router.Listen(ctx)
		if err != nil {
			slog.Error("error listening", slog.Any("message", err))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-messages:
			slog.Debug("handling message", slog.String("routing_key", msg.RoutingKey), slog.String("exchange", msg.Exchange))
			go func() {
				s.Router.Handle(ctx, msg)
			}()
		}
	}
}
