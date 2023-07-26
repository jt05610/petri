package petri

// Arc is a connection from a place to a transition or a transition to a place.
type Arc struct {
	ID string
	// Src is the place or transition that is the source of the arc.
	Src Node
	// Dest is the place or transition that is the destination of the arc.
	Dest Node
}

func (a *Arc) String() string {
	return a.Src.String() + " -> " + a.Dest.String()
}

func NewArc(id string, head, tail Node) *Arc {
	return &Arc{
		ID:   id,
		Src:  head,
		Dest: tail,
	}
}

type ArcInput struct {
	Head string
	Tail string
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

type ArcService interface {
	Service[ArcInput, ArcInput, Arc, ArcFilter]
}

type ArcLoader interface {
	Loader[Arc]
}

type ArcFlusher interface {
	Flusher[Arc]
}
