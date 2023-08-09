package labeled_test

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
	"testing"
)

type printer struct {
	*labeled.Net
	msg interface{}
}

func (p *printer) enqueue(data interface{}) *labeled.Event {
	return &labeled.Event{
		Name: "enqueue",
		Data: data,
	}
}

func (p *printer) enqueued(data interface{}) *labeled.Event {
	return &labeled.Event{
		Name: "enqueued",
		Data: data,
	}
}

func (p *printer) HandleEnqueue(_ context.Context, data *labeled.Event) (*labeled.Event, error) {
	p.msg = data.Data
	return p.enqueued(data.Data), nil
}

func queue() *printer {
	pp := []*petri.Place{
		{ID: "p1", Name: "Q", Bound: 5},
		{Name: "I", ID: "p2"},
		{Name: "B", ID: "p3"},
	}
	tt := []*petri.Transition{
		{Name: "enqueue", ID: "t1"},
		{Name: "start", ID: "t2"},
		{Name: "finish", ID: "t3"},
	}
	pNet := petri.New(pp, tt, []*petri.Arc{
		{Src: tt[0], Dest: pp[0]},
		{Src: pp[0], Dest: tt[1]},
		{Src: pp[1], Dest: tt[1]},
		{Src: tt[1], Dest: pp[2]},
		{Src: pp[2], Dest: tt[2]},
		{Src: tt[2], Dest: pp[1]},
	})

	markedNet := marked.New(pNet, marked.Marking{0, 1, 0})
	net := labeled.New(markedNet)
	p := &printer{}
	err := net.AddHandler("enqueue", tt[0], p.HandleEnqueue)
	if err != nil {
		panic(err)
	}
	err = net.AddNotification("finished", tt[2], func(ctx context.Context) (interface{}, error) {
		return p.msg, nil
	})
	return &printer{
		Net: net,
	}
}
func TestNet_Handle(t *testing.T) {
	q := queue()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := q.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-ch:
				fmt.Println(e)
			}
		}
	}()
	err := q.Handle(ctx, q.enqueue("foo"))
	if err != nil {
		t.Error(err)
	}
}

func TestValidSequence(t *testing.T) {
	q := queue()
	seq := []*labeled.Event{q.enqueue("foo"), q.enqueue("bar"), q.enqueue("baz"), q.enqueue("foo"), q.enqueue("foo")}
	ok := labeled.ValidSequence(q.Net, seq)
	if !ok {
		t.Error("expected valid sequence")
	}
}

type EventData struct {
	Rate   float64 `json:"rate"`
	Volume float64 `json:"volume"`
}

func TestEvent_IsValid(t *testing.T) {
	fields := []*labeled.Field{
		{Name: "rate", Type: "number"},
		{Name: "volume", Type: "number"},
	}

	e := labeled.Event{
		Name:   "pump",
		Other:  "",
		Fields: fields,
		Data: &EventData{
			Rate:   1.0,
			Volume: 2.0,
		},
	}
	if !e.IsValid() {
		t.Error("expected valid event")
	}

	e = labeled.Event{
		Name:   "pump",
		Other:  "",
		Fields: fields,
		Data:   map[string]interface{}{"rate": 1.0, "volume": 2.0},
	}
	if !e.IsValid() {
		t.Error("expected valid event")
	}

	e = labeled.Event{
		Name:   "pump",
		Other:  "",
		Fields: fields,
		Data: struct {
			ValveNumber int `json:"valve_number"`
		}{},
	}
	if e.IsValid() {
		t.Error("expected invalid event")
	}
}
