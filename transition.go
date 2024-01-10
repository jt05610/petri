package petri

import (
	"context"
	"github.com/expr-lang/expr"
)

var _ Object = (*Transition)(nil)
var _ Node = (*Transition)(nil)
var _ Input = (*TransitionInput)(nil)
var _ Update = (*TransitionUpdate)(nil)
var _ Filter = (*TransitionFilter)(nil)

// Transition represents a transition
type Transition struct {
	ID         string `json:"_id"`
	Name       string `json:"name,omitempty"`
	Expression string `json:"expression,omitempty"`
	Handler
	Cold      bool                `json:"cold,omitempty"`
	EventFunc EventFunc[any, any] `json:"-"`
	Event     *EventSchema        `json:"event,omitempty,omitempty"`
}

func (t *Transition) PostInit() error {
	return nil
}

func (t *Transition) Document() Document {
	return Document{
		"_id":        t.ID,
		"name":       t.Name,
		"expression": t.Expression,
		"event":      t.Event,
	}
}

func (t *Transition) IsNode() {}

func (t *Transition) String() string {
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

func (t *Transition) WithEvent(f EventFunc[any, any], url string, input, output *TokenSchema) *Transition {
	t.Cold = true
	t.EventFunc = f
	t.Event = &EventSchema{
		Name:         t.Name,
		Url:          url,
		InputSchema:  *input,
		OutputSchema: *output,
	}
	return t
}

func (t *Transition) Update(u Update) error {
	update, ok := u.(*TransitionUpdate)
	if !ok {
		return ErrWrongUpdate
	}
	if update.Mask.Name {
		t.Name = update.Input.Name
	}
	if update.Mask.Expression {
		t.Expression = update.Input.Expression
	}
	if update.Mask.Event {
		if update.Input.Event == nil {
			t.Event = nil
		} else {
			t.Event = update.Input.Event.Object().(*EventSchema)
		}
	}
	return nil
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

func (t *Transition) WithGenerator(f func(value ...interface{}) ([]*Token[interface{}], error)) *Transition {
	t.Handler = NewGenerator(f)
	return t
}

func (t *Transition) WithTransformer(f func(tokens ...*Token[interface{}]) ([]*Token[interface{}], error)) *Transition {
	t.Handler = NewTransformer(f)
	return t
}

func (t *Transition) CanFire(tokenByType map[string]*Token[interface{}]) bool {
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

type TransitionInput struct {
	Name       string
	Expression string
	Event      *EventInput
}

func (t *TransitionInput) Object() Object {
	return &Transition{
		ID:         ID(),
		Name:       t.Name,
		Expression: t.Expression,
	}
}

func (t *TransitionInput) Kind() Kind {
	return TransitionObject
}

type TransitionMask struct {
	Name       bool
	Expression bool
	Event      bool
}

type TransitionUpdate struct {
	Input *TransitionInput
	Mask  *TransitionMask
}

type TransitionFilter struct {
	ID   *StringSelector `json:"_id,omitempty"`
	Name *StringSelector `json:"name,omitempty"`
}

func (t *TransitionInput) IsInput()       {}
func (t *TransitionUpdate) IsUpdate()     {}
func (t *TransitionFilter) IsFilter()     {}
func (t *TransitionFilter) IsNodeFilter() {}
