package place

import (
	"github.com/google/uuid"
	"pnet"
)

// NewBuilder creates a new place builder. The name is required, and the limit is optional. Without a limit,
// the place is unbounded.
func NewBuilder[T petri.Token](name string, limit ...int) *Builder[T] {
	lim := 0
	if len(limit) > 0 {
		lim = limit[0]
	}

	return &Builder[T]{
		place: Place[T]{
			id:      uuid.New().String(),
			name:    name,
			inputs:  []petri.Node{},
			outputs: []petri.Node{},
			tokens:  petri.NewFIFO[T](lim),
		},
	}
}

func (b *Builder[T]) Build() *Place[T] {
	return &b.place
}

func (b *Builder[T]) AddInputs(n ...petri.Node) *Builder[T] {
	b.place.inputs = append(b.place.inputs, n...)
	return b
}

func (b *Builder[T]) AddOutputs(n ...petri.Node) *Builder[T] {
	b.place.outputs = append(b.place.outputs, n...)
	return b
}

func (b *Builder[T]) AddToken(t T) *Builder[T] {
	b.place.tokens.Push(t)
	return b
}
