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
}

func (t *Transition) IsNode() {}

func (t *Transition) String() string {
	return t.Name
}

func (t *Transition) Kind() Kind { return TransitionObject }

func (t *Transition) Identifier() string { return t.ID }

func (t *Transition) Init(id string, i Input) error {
	in, ok := i.(*TransitionInput)
	if !ok {
		return ErrWrongInput
	}
	t.ID = id
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

func NewTransition(id, name string) *Transition {
	return &Transition{
		ID:   id,
		Name: name,
	}
}

type TransitionInput struct {
	Name string
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

func (t *TransitionInput) IsInput()   {}
func (t *TransitionUpdate) IsUpdate() {}
func (t *TransitionFilter) IsFilter() {}
