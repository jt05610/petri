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
		{Name: "Q", Bound: 5},
		{Name: "I"},
		{Name: "B"},
	}
	tt := []*petri.Transition{
		{Name: "enqueue"},
		{Name: "start"},
		{Name: "finish"},
	}
	pNet := petri.New(pp, tt, []*petri.Arc{
		{Head: tt[0], Tail: pp[0]},
		{Head: pp[0], Tail: tt[1]},
		{Head: pp[1], Tail: tt[1]},
		{Head: tt[1], Tail: pp[2]},
		{Head: pp[2], Tail: tt[2]},
		{Head: tt[2], Tail: pp[1]},
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
