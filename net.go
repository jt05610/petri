package petri

import (
	"errors"
)

// Arc connects places and transitions
type Arc struct {
	Head Node
	Tail Node
}

// Transition represents a transition
type Transition struct {
	Name string
}

func (t *Transition) Kind() NodeKind { return TransitionNode }

// Place represents a place
type Place struct {
	Name string
}

func (p *Place) Kind() NodeKind { return PlaceNode }

type Marking []int

// Net struct
type Net struct {
	Places      []*Place
	Transitions []*Transition
	Arcs        []*Arc
	ArcIndexes  map[Node]map[Node]*Arc
}

func (p *Net) Arc(head, tail Node) *Arc {
	if _, ok := p.ArcIndexes[head]; !ok {
		return nil
	}
	return p.ArcIndexes[head][tail]
}

func (p *Net) Inputs(n Node) []*Arc {
	var inputs []*Arc
	for _, arc := range p.Arcs {
		if arc.Tail == n {
			inputs = append(inputs, arc)
		}
	}
	return inputs
}

func (p *Net) Outputs(n Node) []*Arc {
	var outputs []*Arc
	for _, arc := range p.Arcs {
		if arc.Head == n {
			outputs = append(outputs, arc)
		}
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
	if _, ok := p.ArcIndexes[from]; !ok {
		p.ArcIndexes[from] = make(map[Node]*Arc)
	}
	p.ArcIndexes[from][to] = a
	return a, nil
}

func New(places []*Place, transitions []*Transition, arcs []*Arc) *Net {
	net := &Net{
		Places:      places,
		Transitions: transitions,
		ArcIndexes:  make(map[Node]map[Node]*Arc),
	}
	for _, arc := range arcs {
		if _, ok := net.ArcIndexes[arc.Head]; !ok {
			net.ArcIndexes[arc.Head] = make(map[Node]*Arc)
		}
		net.ArcIndexes[arc.Head][arc.Tail] = arc
	}
	return net
}
