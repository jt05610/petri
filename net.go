package petri

import (
	"errors"
	"io"
)

// Transition represents a transition
type Transition struct {
	ID   string
	Name string
}

func (t *Transition) Kind() NodeKind { return TransitionNode }

func (t *Transition) String() string { return t.Name }

func NewTransition(id, name string) *Transition {
	return &Transition{
		ID:   id,
		Name: name,
	}
}

// Place represents a place
type Place struct {
	ID    string
	Name  string
	Bound int
}

func NewPlace(id, name string, bound int) *Place {
	return &Place{
		ID:    id,
		Name:  name,
		Bound: bound,
	}
}

func (p *Place) Kind() NodeKind { return PlaceNode }

func (p *Place) String() string { return p.Name }

// Net struct
type Net struct {
	ID          string
	Name        string
	Places      []*Place
	Transitions []*Transition
	Arcs        []*Arc
	inputs      map[string][]*Arc
	outputs     map[string][]*Arc
}

func (p *Net) Arc(head, tail Node) *Arc {
	if _, ok := p.outputs[head.String()]; !ok {
		return nil
	}
	for _, arc := range p.outputs[head.String()] {
		if arc.Dest.String() == tail.String() {
			return arc
		}
	}
	return nil
}

func (p *Net) Inputs(n Node) []*Arc {
	var inputs []*Arc
	for _, o := range p.inputs[n.String()] {
		inputs = append(inputs, o)
	}
	return inputs
}

func (p *Net) Outputs(n Node) []*Arc {
	var outputs []*Arc
	for _, o := range p.outputs[n.String()] {
		outputs = append(outputs, o)
	}
	return outputs
}

func (p *Net) AddArc(from, to Node) (*Arc, error) {
	if from.Kind() == to.Kind() {
		return nil, errors.New("cannot connect two places or two transitions")
	}
	if arc := p.Arc(from, to); arc != nil {
		return nil, errors.New("arc already exists")
	}
	a := &Arc{
		Src:  from,
		Dest: to,
	}
	p.Arcs = append(p.Arcs, a)
	if _, ok := p.outputs[from.String()]; !ok {
		p.outputs[from.String()] = make([]*Arc, 0)
	}
	p.outputs[from.String()] = append(p.outputs[from.String()], a)
	if _, ok := p.inputs[to.String()]; !ok {
		p.inputs[to.String()] = make([]*Arc, 0)
	}
	p.inputs[to.String()] = append(p.inputs[to.String()], a)
	return a, nil
}

func New(places []*Place, transitions []*Transition, arcs []*Arc, id ...string) *Net {
	ii := ""
	if len(id) > 0 {
		ii = id[0]
	}
	for _, p := range places {
		if p.Bound == 0 {
			p.Bound = 1
		}
	}
	net := &Net{
		ID:          ii,
		Places:      places,
		Transitions: transitions,
		Arcs:        arcs,
		inputs:      make(map[string][]*Arc),
		outputs:     make(map[string][]*Arc),
	}
	for _, arc := range arcs {
		if _, ok := net.outputs[arc.Src.String()]; !ok {
			net.outputs[arc.Src.String()] = make([]*Arc, 0)
		}
		net.outputs[arc.Src.String()] = append(net.outputs[arc.Src.String()], arc)
		if _, ok := net.inputs[arc.Dest.String()]; !ok {
			net.inputs[arc.Dest.String()] = make([]*Arc, 0)
		}
		net.inputs[arc.Dest.String()] = append(net.inputs[arc.Dest.String()], arc)
	}
	return net
}

type Loader[T any] interface {
	Load(io.Reader) (T, error)
}

type Flusher[T any] interface {
	Flush(io.Writer, T) error
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

type PlaceInput struct {
	Name  string
	Bound int
}

type PlaceMask struct {
	Name  bool
	Bound bool
}

type PlaceFilter struct {
	Name  string
	Bound int
	*PlaceMask
}

type PlaceUpdate struct {
	Input *PlaceInput
	Mask  *PlaceMask
}

type NetInput struct {
	Name        string
	Places      []*PlaceInput
	Transitions []*TransitionInput
	Arcs        []*ArcInput
}

type NetMask struct {
	Name        bool
	Places      bool
	Transitions bool
	Arcs        bool
}

type NetUpdate struct {
	Input *NetInput
	Mask  *NetMask
}

type NetFilter struct {
	ID          string
	Name        string
	Places      []string
	Transitions []string
	Arcs        []string
}

type Getter[T any] interface {
	Get(id string) (*T, error)
}

type Lister[T, F any] interface {
	List(T, F) ([]*T, error)
}

type Adder[T, U any] interface {
	Add(t *T) (*U, error)
}

type Remover[T any] interface {
	Remove(id string) (*T, error)
}

type Updater[T, U any] interface {
	Update(id string, update *T) (*U, error)
}

type Service[I, U, T, F any] interface {
	Getter[T]
	Lister[T, F]
	Adder[I, T]
	Remover[T]
	Updater[U, T]
}

type NetService interface {
	Service[NetInput, NetUpdate, Net, NetFilter]
}

type PlaceService interface {
	Service[PlaceInput, PlaceUpdate, Place, PlaceFilter]
}

type TransitionService interface {
	Service[TransitionInput, TransitionMask, Transition, TransitionFilter]
}
