package petri

var (
	_ Object = (*Arc)(nil)
	_ Input  = (*ArcInput)(nil)
	_ Update = (*ArcUpdate)(nil)
	_ Filter = (*ArcFilter)(nil)
)

// Arc is a connection from a place to a transition or a transition to a place.
type Arc struct {
	ID string
	// Src is the place or transition that is the source of the arc.
	Src Node
	// Dest is the place or transition that is the destination of the arc.
	Dest Node
}

func (a *Arc) Document() Document {
	//TODO implement me
	panic("implement me")
}

func (a *Arc) From(doc Document) error {
	//TODO implement me
	panic("implement me")
}

func (a *Arc) Init(input Input) error {
	in, ok := input.(*ArcInput)
	if !ok {
		return ErrWrongInput
	}
	a.Src = in.Head
	a.Dest = in.Tail
	a.ID = in.ID
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

func NewArc(from, to Node) *Arc {
	return &Arc{
		Src:  from,
		Dest: to,
	}
}

func (a *Arc) Identifier() string {
	return a.ID
}

func (a *Arc) String() string {
	return a.Src.Identifier() + " -> " + a.Dest.Identifier()
}

func (a *Arc) Kind() Kind { return ArcObject }

type ArcInput struct {
	ID   string
	Head Node
	Tail Node
}

func (a *ArcInput) Object() Object {
	//TODO implement me
	panic("implement me")
}

func (a *ArcInput) Kind() Kind {
	return ArcObject
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

func (a *ArcFilter) Filter() Document {
	//TODO implement me
	panic("implement me")
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
