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
	ID         string
	Name       string
	Expression string
	Handler
	Cold  bool
	Event EventFunc[any, any]
}

func (t *Transition) Document() Document {
	//TODO implement me
	panic("implement me")
}

func (t *Transition) From(doc Document) error {
	//TODO implement me
	panic("implement me")
}

func (t *Transition) IsNode() {}

func (t *Transition) String() string {
	return t.Name
}

func (t *Transition) Kind() Kind { return TransitionObject }

func (t *Transition) Identifier() string { return t.ID }

func (t *Transition) Init(i Input) error {
	in, ok := i.(*TransitionInput)
	if !ok {
		return ErrWrongInput
	}
	t.ID = in.ID
	t.Name = in.Name
	return nil
}

type EventFunc[T, U any] func(ctx context.Context, input T) (U, error)

func NewEventFunc[T, U any](f func(ctx context.Context, input T) (U, error)) EventFunc[any, any] {
	return func(ctx context.Context, input any) (any, error) {
		return f(ctx, input.(T))
	}
}

func (t *Transition) WithEvent(f EventFunc[any, any]) *Transition {
	t.Cold = true
	t.Event = f
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
	ID   string
	Name string
}

func (t *TransitionInput) Object() Object {
	//TODO implement me
	panic("implement me")
}

func (t *TransitionInput) Kind() Kind {
	return TransitionObject
}

type TransitionMask struct {
	Name bool
}

type TransitionUpdate struct {
	Input *TransitionInput
	Mask  *TransitionMask
}

type TransitionFilter struct {
	Name string
	*TransitionMask
}

func (t *TransitionFilter) Filter() Document {
	//TODO implement me
	panic("implement me")
}

func (t *TransitionInput) IsInput()   {}
func (t *TransitionUpdate) IsUpdate() {}
func (t *TransitionFilter) IsFilter() {}
