package graph

import (
	"context"
	"pnet/net"
)

type Builder struct {
	places      net.Places
	transitions net.Transitions
	arcs        []*net.Arc
	weights     map[*net.Arc]int
	placeIndex  map[string]int
	transIndex  map[string]int
	arcIndexes  map[net.Node]map[net.Node]*net.Arc
}

func NewBuilder() *Builder {
	return &Builder{
		places:      make(net.Places, 0),
		transitions: make(net.Transitions, 0),
		placeIndex:  make(map[string]int),
		transIndex:  make(map[string]int),
		weights:     make(map[*net.Arc]int),
		arcs:        make([]*net.Arc, 0),
		arcIndexes:  make(map[net.Node]map[net.Node]*net.Arc),
	}
}

func (b *Builder) AddPlace(name string, accepts net.Token, tokens ...net.Token) *Builder {
	b.placeIndex[name] = len(b.places)
	b.places = append(b.places, NewPlace(accepts, tokens...))
	return b
}

func (b *Builder) AddTransition(name string, h func(context.Context, ...net.Token) []net.Token) *Builder {
	b.transIndex[name] = len(b.transitions)
	b.transitions = append(b.transitions, &net.Transition{
		Name:   name,
		Handle: h,
	})
	return b
}

func (b *Builder) AddArc(placeName, tName string, weight int, toTransition bool) *Builder {
	var from net.Node
	var to net.Node
	if toTransition {
		from = b.places[b.placeIndex[placeName]]
		to = b.transitions[b.transIndex[tName]]
	} else {
		from = b.transitions[b.transIndex[tName]]
		to = b.places[b.placeIndex[placeName]]
	}
	arc := &net.Arc{
		Head: from,
		Tail: to,
	}
	b.arcs = append(b.arcs, arc)
	b.weights[arc] = weight
	if _, ok := b.arcIndexes[from]; !ok {
		b.arcIndexes[from] = make(map[net.Node]*net.Arc)
	}

	b.arcIndexes[from][to] = arc
	return b
}

func (b *Builder) Build() *net.PetriNet {
	return &net.PetriNet{
		Places:      b.places,
		Transitions: b.transitions,
		Arcs:        b.arcs,
		Weights:     b.weights,
		ArcIndexes:  b.arcIndexes,
	}
}
