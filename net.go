package petri

import (
	"errors"
	"io"
)

// Arc connects places and transitions
type Arc struct {
	Head Node
	Tail Node
}

func (a *Arc) String() string {
	return a.Head.String() + " -> " + a.Tail.String()
}

// Transition represents a transition
type Transition struct {
	Name string
}

func (t *Transition) Kind() NodeKind { return TransitionNode }

func (t *Transition) String() string { return t.Name }

// Place represents a place
type Place struct {
	Name  string
	Bound int
}

func (p *Place) Kind() NodeKind { return PlaceNode }

func (p *Place) String() string { return p.Name }

// Net struct
type Net struct {
	Places      []*Place
	Transitions []*Transition
	Arcs        []*Arc
	inputs      map[string][]*Arc
	outputs     map[string][]*Arc
}

func (p *Net) Arc(head, tail Node) *Arc {
	if _, ok := p.outputs[head.String()]; !ok {
		return nil
	}
	for _, arc := range p.outputs[head.String()] {
		if arc.Tail.String() == tail.String() {
			return arc
		}
	}
	return nil
}

func (p *Net) Inputs(n Node) []*Arc {
	var inputs []*Arc
	for _, o := range p.inputs[n.String()] {
		inputs = append(inputs, o)
	}
	return inputs
}

func (p *Net) Outputs(n Node) []*Arc {
	var outputs []*Arc
	for _, o := range p.outputs[n.String()] {
		outputs = append(outputs, o)
	}
	return outputs
}

func (p *Net) AddArc(from, to Node) (*Arc, error) {
	if from.Kind() == to.Kind() {
		return nil, errors.New("cannot connect two places or two transitions")
	}
	if arc := p.Arc(from, to); arc != nil {
		return nil, errors.New("arc already exists")
	}
	a := &Arc{
		Head: from,
		Tail: to,
	}
	p.Arcs = append(p.Arcs, a)
	if _, ok := p.outputs[from.String()]; !ok {
		p.outputs[from.String()] = make([]*Arc, 0)
	}
	p.outputs[from.String()] = append(p.outputs[from.String()], a)
	if _, ok := p.inputs[to.String()]; !ok {
		p.inputs[to.String()] = make([]*Arc, 0)
	}
	p.inputs[to.String()] = append(p.inputs[to.String()], a)
	return a, nil
}

func New(places []*Place, transitions []*Transition, arcs []*Arc) *Net {
	for _, p := range places {
		if p.Bound == 0 {
			p.Bound = 1
		}
	}
	net := &Net{
		Places:      places,
		Transitions: transitions,
		Arcs:        arcs,
		inputs:      make(map[string][]*Arc),
		outputs:     make(map[string][]*Arc),
	}
	for _, arc := range arcs {
		if _, ok := net.outputs[arc.Head.String()]; !ok {
			net.outputs[arc.Head.String()] = make([]*Arc, 0)
		}
		net.outputs[arc.Head.String()] = append(net.outputs[arc.Head.String()], arc)
		if _, ok := net.inputs[arc.Tail.String()]; !ok {
			net.inputs[arc.Tail.String()] = make([]*Arc, 0)
		}
		net.inputs[arc.Tail.String()] = append(net.inputs[arc.Tail.String()], arc)
	}
	return net
}

type NetService interface {
	Load(io.Reader) (*Net, error)
	Flush(io.Writer, *Net) error
}
