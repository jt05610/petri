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

func (t *Transition) String() string { return t.Name }

// Place represents a place
type Place struct {
	Name string
}

func (p *Place) Kind() NodeKind { return PlaceNode }

func (p *Place) String() string { return p.Name }

type Marking []int

// Net struct
type Net struct {
	Places      []*Place
	Transitions []*Transition
	Arcs        []*Arc
	ArcIndexes  map[string]map[string]*Arc
}

func (p *Net) Arc(head, tail Node) *Arc {
	if _, ok := p.ArcIndexes[head.String()]; !ok {
		return nil
	}
	return p.ArcIndexes[head.String()][tail.String()]
}

func (p *Net) Inputs(n Node) []*Arc {
	var inputs []*Arc
	for _, o := range p.Arcs {
		if o.Tail.String() == n.String() {
			inputs = append(inputs, o)
		}
	}
	return inputs
}

func (p *Net) Outputs(n Node) []*Arc {
	var outputs []*Arc
	for _, o := range p.ArcIndexes[n.String()] {
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
	if _, ok := p.ArcIndexes[from.String()]; !ok {
		p.ArcIndexes[from.String()] = make(map[string]*Arc)
	}
	p.ArcIndexes[from.String()][to.String()] = a
	return a, nil
}

func New(places []*Place, transitions []*Transition, arcs []*Arc) *Net {
	net := &Net{
		Places:      places,
		Transitions: transitions,
		Arcs:        arcs,
		ArcIndexes:  make(map[string]map[string]*Arc),
	}
	for _, arc := range arcs {
		if _, ok := net.ArcIndexes[arc.Head.String()]; !ok {
			net.ArcIndexes[arc.Head.String()] = make(map[string]*Arc)
		}
		net.ArcIndexes[arc.Head.String()][arc.Tail.String()] = arc
	}
	return net
}
