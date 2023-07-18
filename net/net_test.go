package net_test

import (
	"context"
	"pnet/net"
	"pnet/net/graph"
	"testing"
)

type Tok int

func (t Tok) IsToken() {}

func CaseNet() *net.PetriNet {
	builder := graph.NewBuilder().
		AddPlace("p1", Tok(1), Tok(1), Tok(2)).
		AddPlace("p2", Tok(0)).
		AddPlace("p3", Tok(0)).
		AddPlace("p4", Tok(2), Tok(2)).
		AddTransition("t1", func(_ context.Context, i ...net.Token) []net.Token {
			return []net.Token{Tok(1), Tok(2)}
		}).
		AddTransition("t2", func(_ context.Context, i ...net.Token) []net.Token {
			return i
		}).
		AddTransition("t3", func(_ context.Context, i ...net.Token) []net.Token {
			return nil
		}).
		AddArc("p1", "t1", 1, true).
		AddArc("p1", "t3", 1, true).
		AddArc("p2", "t1", 1, false).
		AddArc("p2", "t2", 1, false).
		AddArc("p2", "t2", 1, true).
		AddArc("p3", "t1", 1, false).
		AddArc("p3", "t2", 1, true).
		AddArc("p3", "t3", 1, true).
		AddArc("p4", "t2", 1, false).
		AddArc("p4", "t3", 1, true)

	return builder.Build()

}

type Customer struct {
	Problem string
}

func (c *Customer) IsToken() {}

type Server struct {
	c *Customer
}

func (s *Server) IsToken() {}

func QueueCase() *net.PetriNet {
	builder := graph.NewBuilder().
		AddPlace("A", &Customer{}, &Customer{}).
		AddPlace("Q", &Customer{}).
		AddPlace("I", &Server{}).
		AddPlace("B", &Server{}).
		AddPlace("F", &Customer{}).
		AddTransition("arrives", func(_ context.Context, i ...net.Token) []net.Token {
			return i
		}).
		AddTransition("starts", func(_ context.Context, i ...net.Token) []net.Token {
			var customer *Customer
			var server *Server
			for _, token := range i {
				switch t := token.(type) {
				case *Customer:
					customer = t
				case *Server:
					server = t
				}
			}
			if customer == nil || server == nil {
				panic("missing customer or server")
			}
			server.c = customer
			return []net.Token{server}
		}).
		AddTransition("finishes", func(_ context.Context, i ...net.Token) []net.Token {
			var server *Server
			for _, token := range i {
				switch t := token.(type) {
				case *Server:
					server = t
				}
			}
			if server == nil {
				panic("missing server")
			}
			c := server.c
			server.c = nil
			return []net.Token{server, c}
		}).
		AddTransition("leaves", func(_ context.Context, i ...net.Token) []net.Token {
			return nil
		}).
		AddArc("A", "arrives", 1, true).
		AddArc("A", "arrives", 1, false).
		AddArc("Q", "arrives", 1, false).
		AddArc("Q", "starts", 1, true).
		AddArc("I", "starts", 1, true).
		AddArc("B", "starts", 1, false).
		AddArc("B", "finishes", 1, true).
		AddArc("I", "finishes", 1, false).
		AddArc("F", "finishes", 1, false).
		AddArc("F", "leaves", 1, true)

	return builder.Build()
}

func TestPetriNet_Enabled(t *testing.T) {
	petri := CaseNet()
	if len(petri.Available()) != 1 {
		t.Errorf("should only be one available transition")
	}
	if !petri.Enabled(petri.Transitions[0]) {
		t.Errorf("t1 should be enabled")
	}
	mark := petri.State()
	expected := []float64{
		2,
		0,
		0,
		1,
	}
	for name, count := range expected {
		if mark[name] != count {
			t.Errorf("expected %f tokens in %d, got %f", count, name, mark[name])
		}
	}

	ctx := context.Background()
	err := petri.Fire(ctx, 0)
	if err != nil {
		t.Errorf("should be no error firing t1")
	}
	expected = []float64{
		1,
		1,
		1,
		1,
	}
	mark = petri.State()
	for name, count := range expected {
		if mark[name] != count {
			t.Errorf("expected %f tokens in %d, got %f", count, name, mark[name])
		}
	}
	available := petri.Available()
	if len(available) != 3 {
		t.Errorf("should be 3 available transitions")
	}
	err = petri.Fire(ctx, 1)
	if err != nil {
		t.Errorf("should be no error firing t2")
	}
	expected = []float64{
		1,
		1,
		0,
		2,
	}
	mark = petri.State()
	for name, count := range expected {
		if mark[name] != count {
			t.Errorf("expected %f tokens in %d, got %f", count, name, mark[name])
		}
	}
	if len(petri.Available()) != 1 {
		t.Errorf("should be 2 available transitions")
	}
}

func TestPetriNet_Incidence(t *testing.T) {
	petri := CaseNet()
	expected := [][]float64{
		{-1, 1, 1, 0},
		{0, 0, -1, 1},
		{-1, 0, -1, -1},
	}
	inc := petri.Incidence()
	r, c := inc.Dims()
	if r != 3 {
		t.Errorf("should be 3 rows in incidence matrix")
	}
	if c != 4 {
		t.Errorf("should be 4 columns in incidence matrix")
	}
	for i, row := range expected {
		for j, v := range row {
			if int(inc.At(i, j)) != int(v) {
				t.Errorf("expected %f at %d,%d, got %f", v, i, j, inc.At(i, j))
			}
		}
	}

	initial := &net.State{2, 0, 0, 1}
	out, valid := petri.NextState(initial, petri.Transitions[0])
	if !valid {
		t.Errorf("should be valid")
	}
	firedExpect := []float64{1, 1, 1, 1}
	for i, v := range firedExpect {
		if (*out)[i] != v {
			t.Errorf("expected %f at %d, got %f", v, i, (*out)[i])
		}
	}
	out, valid = petri.NextState(out, petri.Transitions[1])
	firedExpect = []float64{1, 1, 0, 2}
	for i, v := range firedExpect {
		if (*out)[i] != v {
			t.Errorf("expected %f at %d, got %f", v, i, (*out)[i])
		}
	}
}

func BenchmarkPetriNet_Incidence(b *testing.B) {
	petri := CaseNet()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		petri.Incidence()
	}
}

func TestQueue(t *testing.T) {
	queue := QueueCase()
	initial := &net.State{1, 0, 1, 0, 0}
	sequence := []*net.Transition{
		queue.Transitions[0],
		queue.Transitions[1],
		queue.Transitions[0],
		queue.Transitions[0],
		queue.Transitions[2],
		queue.Transitions[1],
		queue.Transitions[3],
		queue.Transitions[0],
	}
	next := initial
	var valid bool
	for _, transition := range sequence {
		next, valid = queue.NextState(next, transition)
		if !valid {
			t.Errorf("should be valid")
		}
	}
	expect := &net.State{1, 2, 0, 1, 0}
	for i, v := range *next {
		if v != (*expect)[i] {
			t.Errorf("expected %f at %d, got %f", (*expect)[i], i, v)
		}
	}
	if !queue.Reachable(initial, expect) {
		t.Errorf("should be reachable")
	}
}
