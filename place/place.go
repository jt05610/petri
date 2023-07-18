package place

import (
	"context"
	"errors"
)

var _ petri.Node = &Place[petri.Token]{}

var (
	ErrPlaceFull       = errors.New("place is full")
	ErrNotEnoughTokens = errors.New("not enough tokens")
)

// Place is a node that holds information related to Transition conditions. Places can be inputs to transitions or
// outputs from transitions. Places can have multiple inputs and outputs, but are restricted to one kind of Token.
// A Place can be bound, meaning that it can only hold a certain number of tokens. If a Place is bound, then it will
// not accept any more tokens if it is full and will instead return an error.
type Place[T petri.Token] struct {
	// id is a unique identifier for the Place
	id string
	// name is a human-readable name for the Place
	name string
	// bound is the maximum number of tokens that the Place can hold. If 0, then the Place is unbound.
	bound int
	// tokens is a FIFO queue of tokens currently in the Place.
	tokens *petri.FIFO[T]
	// inputs is a list of nodes that are inputs to the Place.
	inputs []petri.Node
	// outputs is a list of nodes that are outputs from the Place.
	outputs []petri.Node
}

func (p *Place[T]) Inputs(_ context.Context) []petri.Node {
	return p.inputs
}

func (p *Place[T]) Outputs(_ context.Context) []petri.Node {
	return p.outputs
}

func (p *Place[T]) ID() string {
	return p.id
}

func (p *Place[T]) Name() string {
	return p.name
}

func (p *Place[T]) Add(t ...T) error {
	if p.bound > 0 && p.tokens.Len() >= p.bound {
		return ErrPlaceFull
	}
	p.tokens.Push(t...)
	return nil
}

func (p *Place[T]) Pop(n int) ([]T, error) {
	if n > p.tokens.Len() {
		return nil, ErrNotEnoughTokens
	}
	return p.tokens.Pop(n), nil
}

type Builder[T petri.Token] struct {
	place Place[T]
}
