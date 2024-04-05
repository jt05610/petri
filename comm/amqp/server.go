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
	ch       chan<- Update
	deferral func()
}

type MarkingUpdate struct {
	PlaceID string
	Tokens  []petri.Token
}

func NewQueue(placeID string, q petri.TokenQueue, deferral func(), signal chan<- Update) *Queue {
	return &Queue{
		TokenQueue: q,
		ch:         signal,
		PlaceID:    placeID,
		deferral:   deferral,
	}
}

func (q *Queue) Close() {
	q.TokenQueue.Close()
	if q.cancel != nil {
		q.cancel()
	}
	if q.deferral != nil {
		q.deferral()
	}
}

type RPC struct {
	amqp.Delivery
}

type Bindings map[string]HandlerFunc

type Update struct {
	PlaceID string
	Tokens  []petri.Token
}

type Router struct {
	marking petri.Marking
	mu      sync.Mutex
	Bindings
	*petri.Net
	*amqp.Channel
	Exchange    string
	Name        string
	q           amqp.Queue
	Timeout     time.Duration
	ch          chan Update
	ctx         context.Context
	conn        *amqp.Connection
	connections map[string]*amqp.Connection
	cancel      func()
	queues      map[string]*Queue
}

func SameBytes(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i, b := range b1 {
		if b != b2[i] {
			return false
		}
	}
	return true
}

func SameMarking(m1, m2 map[string][]petri.Token) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if len(v) != len(m2[k]) {
			return false
		}
		for i, t := range v {
			if !SameBytes(t.Bytes(), m2[k][i].Bytes()) {
				return false
			}
		}
	}
	return true
}

func (r *Router) State() petri.Marking {
	slog.Info("trying for lock")
	r.mu.Lock()
	slog.Info("locking state")
	defer func() {
		r.mu.Unlock()
		slog.Info("freeing state")
	}()
	return r.marking
}

func (r *Router) Close() {
	for _, conn := range r.connections {
		err := conn.Close()
		if err != nil {
			slog.Error("error closing amqp connection", slog.Any("message", err))
		}
	}
	r.cancel()
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
		slog.Info("bound route", slog.String("route", route), slog.String("routing_key", fmt.Sprintf("%s.%s", r.Exchange, route)))
	}
	return nil
}

func (r *Router) bindEvent(name string, tr *petri.Transition) error {
	pl := petri.NewPlace(name, 1, tr.EventSchema)
	pl.ID = name
	r.Net = r.Net.WithPlaces(pl)
	r.Net = r.Net.WithArcs(petri.NewArc(pl, tr, tr.EventSchema.Name, tr.EventSchema))
	return nil
}

func NewRouter(n *petri.Net, q amqp.Queue, conn *amqp.Connection, initial petri.Marking, name, exchange string) (*Router, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	r := &Router{
		Channel:     ch,
		Exchange:    exchange,
		Name:        name,
		marking:     initial,
		ch:          make(chan Update, 100),
		Net:         n,
		q:           q,
		conn:        conn,
		connections: make(map[string]*amqp.Connection),
		queues:      make(map[string]*Queue),
	}
	r.Bindings = r.MakeBindings()
	for tn, t := range n.Transitions {
		if !t.Cold {
			continue
		}
		err := r.bindEvent(tn, t)
		if err != nil {
			return nil, err
		}
	}
	err = r.ExposePlaces(r.Net)
	if err != nil {
		return nil, err
	}
	r.InitializeMarking()
	return r, nil
}

func (r *Router) Mark(marking petri.Marking) {
	r.mu.Lock()
	slog.Info("locking marking", slog.Any("new marking", marking))
	defer func() {
		r.mu.Unlock()
		slog.Info("freeing marking", slog.Any("new marking", marking))
	}()
	r.marking = marking
}

func (r *Router) InitializeMarking() {
	mark := r.Net.NewMarking()
	for id, q := range r.State() {
		for _, tok := range q {
			err := mark[id].Enqueue(r.ctx, tok)
			if err != nil {
				slog.Error("error enqueuing token", slog.Any("message", err))
			}
		}
	}
	r.Mark(mark.Mark())
	slog.Info("initialized marking", slog.Any("marking", r.State()))
}

func (r *Router) Listen(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)
	defer r.cancel()
	defer close(r.ch)
	for _, q := range r.queues {
		go func(q *Queue) {
			defer q.Close()
			msgs := q.Monitor(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case tt := <-msgs:
					slog.Info("queue updated", slog.String("place", q.PlaceID))
					r.ch <- Update{
						PlaceID: q.PlaceID,
						Tokens:  tt,
					}
				}
			}
		}(q)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-r.ch:
			slog.Info("marking updated", slog.Any("update", update))
			state := r.State()
			state[update.PlaceID] = update.Tokens
			slog.Info("state", slog.Any("state", state))
			updated, err := r.Net.Process(state)
			if err != nil {
				if !errors.Is(petri.ErrNoEvents, err) {
					slog.Error("error processing marking", slog.Any("message", err))
				}
			}
			r.Mark(updated)
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
	None
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

var PutCodec = RPCCodec[PutRequest, Response]{
	Request:  JSONCodec[PutRequest]{},
	Response: JSONCodec[Response]{},
}

var PopCodec = RPCCodec[PopRequest, PopResponse]{
	Request:  JSONCodec[PopRequest]{},
	Response: JSONCodec[PopResponse]{},
}

// DisconnectRequest is a request to disconnect a transition from a place on another petri net.
type DisconnectRequest struct {
	Url       string
	Exchange  string
	PlaceName string
}

// HandleConnect connects a transition to a place on another petri net using an amqp.Remote queue. It begins monitoring the queue for changes in the marking of the place.
func (r *Router) HandleConnect(_ context.Context, request *ConnectRequest) (*Response, error) {
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
	pl := r.Place(request.PlaceName)
	if pl == nil {
		return &Response{
			Ok:    false,
			Error: "no place found",
		}, nil
	}
	// now we need to connect to the remote place on the other petri net
	conn, err := amqp.Dial(request.Url)
	if err != nil {
		return &Response{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	r.connections[request.Url] = conn
	q := queue.NewRemote(request.Exchange, conn, pl)
	d := func() {
		err := conn.Close()
		if err != nil {
			slog.Error("error closing amqp connection", slog.Any("message", err))
		}
	}
	rem := NewQueue(pl.ID, q, d, r.ch)
	r.queues[pl.ID] = rem
	if r.ctx == nil {
		return &Response{
			Ok: true,
		}, nil
	}
	return &Response{
		Ok: true,
	}, nil
}

// HandleDisconnect disconnects a transition from a place on another petri net. It stops monitoring the queue for
// changes in the marking of the place, and closes any connections that were made in connecting to the place.
func (r *Router) HandleDisconnect(_ context.Context, request *DisconnectRequest) (*Response, error) {
	pl := r.Place(request.PlaceName)
	if pl == nil {
		return &Response{
			Ok:    false,
			Error: "no place found",
		}, nil
	}
	state := r.State()
	q := state[request.PlaceName]
	if q == nil {
		return &Response{
			Ok:    false,
			Error: "no queue found",
		}, nil
	}
	delete(state, request.PlaceName)
	for _, input := range r.Inputs(pl) {
		r.Net = r.Net.WithoutArc(input)
	}
	for _, output := range r.Outputs(pl) {
		r.Net = r.Net.WithoutArc(output)
	}
	r.Net = r.Net.WithoutPlace(pl)
	delete(state, pl.ID)
	r.Mark(state)
	return &Response{
		Ok: true,
	}, nil
}

func (r *Router) handlePut(ctx context.Context, request *PutRequest) (*Response, error) {
	pl := r.Place(request.Place)
	if pl == nil {
		return &Response{
			Ok:    false,
			Error: "no place found",
		}, nil
	}
	q := pl.TokenQueue
	if q == nil {
		return &Response{
			Ok:    false,
			Error: "no queue found",
		}, nil
	}
	schema := pl.AcceptedTokens[0]
	tok, err := schema.NewToken(request.Token)
	if err != nil {
		return &Response{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	err = q.Enqueue(ctx, tok)
	if err != nil {
		return &Response{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	return &Response{
		Ok: true,
	}, nil
}

func (r *Router) handlePop(ctx context.Context, request *PopRequest) (*PopResponse, error) {
	pl := r.Place(request.Place)
	if pl == nil {
		return nil, fmt.Errorf("no place named %s", request.Place)
	}
	q := pl.TokenQueue
	if q == nil {
		return nil, fmt.Errorf("no queue found")
	}
	t, err := q.Dequeue(ctx)
	if err != nil {
		return nil, err
	}
	return &PopResponse{
		Place: request.Place,
		Token: t,
	}, nil
}

func (r *Router) MakeBindings() Bindings {
	return Bindings{
		"connect":    NewHandler(r.Net.Name, ConnectCodec, r.HandleConnect),
		"disconnect": NewHandler(r.Net.Name, DisconnectCodec, r.HandleDisconnect),
		"put":        NewHandler(r.Net.Name, PutCodec, r.handlePut),
		"pop":        NewHandler(r.Net.Name, PopCodec, r.handlePop),
	}
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
	q        amqp.Queue
}

func (s *Server) Close() {
	err := s.ch.Close()
	if err != nil {
		slog.Error("error closing amqp channel", slog.Any("message", err))
	}
}

func (r *Router) ExposePlaces(n *petri.Net) error {
	for k, pl := range n.Places {
		slog.Info("exposing place", slog.String("name", k), slog.String("exchange", r.Net.Name))
		local := queue.NewLocal(r.Net.Name, r.conn, pl)
		r.queues[pl.ID] = NewQueue(pl.ID, local, func() {
		}, r.ch)
		pl.TokenQueue = r.queues[pl.ID].TokenQueue
	}
	return nil
}

func (r *Router) ConnectPlaces(n *petri.Net, url string) error {
	for k, pl := range n.Places {
		slog.Info("connecting remote place", slog.String("name", k), slog.String("exchange", n.Name), slog.String("url", url))
		_, found := r.queues[pl.ID]
		if found {
			r.queues[pl.ID].Close()
		}
		conn, err := amqp.Dial(url)
		if err != nil {
			return err
		}
		r.queues[pl.ID] = NewQueue(k, queue.NewRemote(k, conn, pl), func() {
			err := conn.Close()
			if err != nil {
				slog.Error("error closing amqp connection", slog.Any("message", err))
			}
		}, r.ch)
		r.connections[url] = conn
		pl.TokenQueue = r.queues[pl.ID].TokenQueue
	}
	return nil
}

func NewServer(conn *amqp.Connection, n *petri.Net, initial petri.Marking) *Server {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	err = ch.ExchangeDeclare(n.Name, "topic", true, false, false, false, nil)
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
	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		panic(err)
	}
	s := &Server{
		Net:     n,
		conn:    conn,
		initial: initial,
		ch:      ch,
		q:       q,
		Timeout: 30 * time.Second,
	}

	s.Router, err = NewRouter(s.Net, q, s.conn, s.initial, q.Name, s.Net.Name)
	if err != nil {
		panic(err)
	}
	s.Router.Net = s.Net
	s.Router.Timeout = s.Timeout
	err = s.Router.Bind()
	if err != nil {
		panic(err)
	}
	return s
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
	defer func() {
		err := s.ch.Close()
		if err != nil {
			slog.Error("error closing amqp channel", slog.Any("message", err))
		}
	}()
	for _, q := range s.queues {
		switch pub := q.TokenQueue.(type) {
		case *queue.Public:
			go func(pub *queue.Public) {
				slog.Info("serving public queue", slog.String("name", pub.Name))
				err := pub.Serve(ctx)
				if err != nil {
					slog.Error("error serving public queue", slog.Any("message", err))
				}
			}(pub)
		}
	}
	messages, err := s.ch.Consume(s.q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
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
			slog.Info("handling message", slog.String("routing_key", msg.RoutingKey), slog.String("exchange", msg.Exchange))
			go func() {
				s.Router.Handle(ctx, msg)
				slog.Info("handled message", slog.String("routing_key", msg.RoutingKey), slog.String("exchange", msg.Exchange))
			}()
		}
	}
}
