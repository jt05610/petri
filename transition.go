package petri

var _ Object = (*Transition)(nil)
var _ Node = (*Transition)(nil)
var _ Input = (*TransitionInput)(nil)
var _ Update = (*TransitionUpdate)(nil)
var _ Filter = (*TransitionFilter)(nil)

// Transition represents a transition
type Transition struct {
	ID   string
	Name string
	Handler
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

func NewTransition(name string) *Transition {
	return &Transition{
		Name: name,
	}
}

func (t *Transition) WithHandler(h Handler) *Transition {
	t.Handler = h
	return t
}

func (t *Transition) WithGenerator(f func(value interface{}) (*Token, error)) *Transition {
	t.Handler = NewGenerator(f)
	return t
}

func (t *Transition) WithTransformer(f func(t *Token) (*Token, error)) *Transition {
	t.Handler = NewTransformer(f)
	return t
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
