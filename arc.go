package petri

var _ Object = (*Arc)(nil)
var _ Input = (*ArcInput)(nil)
var _ Update = (*ArcUpdate)(nil)
var _ Filter = (*ArcFilter)(nil)

// Arc is a connection from a place to a transition or a transition to a place.
type Arc struct {
	ID string
	// Src is the place or transition that is the source of the arc.
	Src Node
	// Dest is the place or transition that is the destination of the arc.
	Dest Node
}

func (a *Arc) Init(id string, input Input) error {
	in, ok := input.(*ArcInput)
	if !ok {
		return ErrWrongInput
	}
	a.Src = in.Head
	a.Dest = in.Tail
	a.ID = id
	return nil
}

func (a *Arc) Update(update Update) error {
	up, ok := update.(*ArcUpdate)
	if !ok {
		return ErrWrongUpdate
	}
	if up.Mask.Head {
		a.Src = up.Input.Head
	}
	if up.Mask.Tail {
		a.Dest = up.Input.Tail
	}
	return nil
}

func (a *Arc) Identifier() string {
	return a.ID
}

func (a *Arc) String() string {
	return a.Src.Identifier() + " -> " + a.Dest.Identifier()
}

func NewArc(id string, head, tail Node) *Arc {
	return &Arc{
		ID:   id,
		Src:  head,
		Dest: tail,
	}
}

func (a *Arc) Kind() Kind { return ArcObject }

type ArcInput struct {
	Head Node
	Tail Node
}

type ArcMask struct {
	Head bool
	Tail bool
}

type ArcUpdate struct {
	Input ArcInput
	Mask  *ArcMask
}

type ArcFilter struct {
	Head string
	Tail string
	*ArcMask
}

type ArcLoader interface {
	Loader[*Arc]
}

type ArcFlusher interface {
	Flusher[*Arc]
}

func (a *ArcInput) IsInput()   {}
func (a *ArcUpdate) IsUpdate() {}
func (a *ArcFilter) IsFilter() {}
