package petri

import (
	"context"
	"github.com/expr-lang/expr"
)

var _ Node = (*Transition)(nil)

// Transition represents a transition
type Transition struct {
	ID         string `json:"_id"`
	Name       string `json:"name,omitempty"`
	Expression string `json:"expression,omitempty"`
	Handler
	Cold        bool         `json:"cold,omitempty"`
	EventSchema *TokenSchema `json:"inputSchema"`
}

func (t *Transition) PostInit() error {
	return nil
}

func (t *Transition) IsNode() {}

func (t *Transition) String() string {
	return t.Name
}

func (t *Transition) Index() string {
	return t.Name
}

func (t *Transition) Kind() Kind { return TransitionObject }

func (t *Transition) Identifier() string { return t.ID }

type EventFunc[T, U any] func(ctx context.Context, input T) (U, error)

func NewEventFunc[T, U any](f func(ctx context.Context, input T) (U, error)) EventFunc[any, any] {
	return func(ctx context.Context, input any) (any, error) {
		return f(ctx, input.(T))
	}
}

func (t *Transition) WithEvent(schema *TokenSchema) *Transition {
	t.Cold = true
	t.EventSchema = schema
	return t
}

func NewTransition(name string, expression ...string) *Transition {
	if len(expression) > 0 {
		return &Transition{
			ID:         ID(),
			Name:       name,
			Expression: expression[0],
		}
	}
	return &Transition{
		ID:   ID(),
		Name: name,
	}
}

func (t *Transition) WithHandler(h Handler) *Transition {
	t.Handler = h
	return t
}

func (t *Transition) CanFire(tokenByType map[string]Token) bool {
	if t.Expression == "" {
		return true
	}
	program, err := expr.Compile(t.Expression)
	if err != nil {
		panic(err)
	}
	ret, err := expr.Run(program, ToValueMap(tokenByType))
	if err != nil {
		panic(err)
	}
	return ret.(bool)
}

type TransitionMask struct {
	Name       bool
	Expression bool
	Event      bool
}
