package petri

import "context"

type Node interface {
	ID() string
	Name() string
	Inputs(ctx context.Context) []Node
	Outputs(ctx context.Context) []Node
}
