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
	msg map[string]interface{}
}

func (p *printer) enqueue(data map[string]interface{}) *labeled.Event {
	return &labeled.Event{
		Name: "enqueue",
		Data: data,
	}
}

func (p *printer) enqueued(data map[string]interface{}) *labeled.Event {
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
	err := net.AddHandler("enqueue", "enqueue", tt[0], p.HandleEnqueue)
	if err != nil {
		panic(err)
	}
	err = net.AddNotification("finished", tt[2], func(ctx context.Context) (map[string]interface{}, error) {
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
	err := q.Handle(ctx, q.enqueue(map[string]interface{}{"foo": "bar"}))
	if err != nil {
		t.Error(err)
	}
}

func TestValidSequence(t *testing.T) {
	q := queue()
	seq := []*labeled.Event{
		q.enqueue(map[string]interface{}{"foo": "bar"}),
		q.enqueue(map[string]interface{}{"bar": "food"}),
		q.enqueue(map[string]interface{}{"foo": "bar"}),
		q.enqueue(map[string]interface{}{"bar": "food"}),
		q.enqueue(map[string]interface{}{"foo": "bar"}),
	}
	ok := labeled.ValidSequence(q.Net, seq)
	if !ok {
		t.Error("expected valid sequence")
	}
}

func TestEvent_IsValid(t *testing.T) {
	fields := []*labeled.Field{
		{Name: "rate", Type: "number"},
		{Name: "volume", Type: "number"},
	}

	e := labeled.Event{
		Name:   "pump",
		Fields: fields,
		Data: map[string]interface{}{
			"rate":   1.0,
			"volume": 2.0,
		},
	}
	if !e.IsValid() {
		t.Error("expected valid event")
	}

	e = labeled.Event{
		Name:   "pump",
		Fields: fields,
		Data:   map[string]interface{}{"rate": 1.0, "volume": 2.0},
	}
	if !e.IsValid() {
		t.Error("expected valid event")
	}

	e = labeled.Event{
		Name:   "pump",
		Fields: fields,
		Data: map[string]interface{}{
			"distance": 5,
		},
	}
	//  should only accept rate and volume
	if e.IsValid() {
		t.Error("expected invalid event")
	}
}
