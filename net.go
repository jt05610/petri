package petri

import (
	"errors"
	"io"
)

var ErrWrongInput = errors.New("wrong input")
var ErrWrongUpdate = errors.New("wrong update")

var _ Object = (*Net)(nil)
var _ Input = (*NetInput)(nil)
var _ Update = (*NetUpdate)(nil)
var _ Filter = (*NetFilter)(nil)

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

func (p *Net) Init(id string, input Input) error {
	in, ok := input.(*NetInput)
	if !ok {
		return ErrWrongInput
	}
	p.ID = id
	p.Name = in.Name
	p.Places = in.Places
	p.Transitions = in.Transitions
	p.Arcs = in.Arcs
	p.inputs = make(map[string][]*Arc)
	p.outputs = make(map[string][]*Arc)
	for _, arc := range p.Arcs {
		if _, ok := p.outputs[arc.Src.Identifier()]; !ok {
			p.outputs[arc.Src.Identifier()] = make([]*Arc, 0)
		}
		p.outputs[arc.Src.Identifier()] = append(p.outputs[arc.Src.Identifier()], arc)
		if _, ok := p.inputs[arc.Dest.Identifier()]; !ok {
			p.inputs[arc.Dest.Identifier()] = make([]*Arc, 0)
		}
		p.inputs[arc.Dest.Identifier()] = append(p.inputs[arc.Dest.Identifier()], arc)
	}
	return nil
}

func (p *Net) Update(update Update) error {
	u, ok := update.(*NetUpdate)
	if !ok {
		return ErrWrongUpdate
	}
	if u.Mask.Name {
		p.Name = u.Input.Name
	}
	if u.Mask.Places {
		p.Places = u.Input.Places
	}
	if u.Mask.Transitions {
		p.Transitions = u.Input.Transitions
	}
	if u.Mask.Arcs {
		p.Arcs = u.Input.Arcs
	}
	p.inputs = make(map[string][]*Arc)
	p.outputs = make(map[string][]*Arc)
	for _, arc := range p.Arcs {
		if _, ok := p.outputs[arc.Src.Identifier()]; !ok {
			p.outputs[arc.Src.Identifier()] = make([]*Arc, 0)
		}
		p.outputs[arc.Src.Identifier()] = append(p.outputs[arc.Src.Identifier()], arc)
		if _, ok := p.inputs[arc.Dest.Identifier()]; !ok {
			p.inputs[arc.Dest.Identifier()] = make([]*Arc, 0)
		}
		p.inputs[arc.Dest.Identifier()] = append(p.inputs[arc.Dest.Identifier()], arc)
	}
	return nil
}

func (p *Net) Identifier() string {
	return p.ID
}

func (p *Net) String() string {
	return p.Name
}

func (p *Net) Arc(head, tail Node) *Arc {
	if _, ok := p.outputs[head.Identifier()]; !ok {
		return nil
	}
	for _, arc := range p.outputs[head.Identifier()] {
		if arc.Dest.Identifier() == tail.Identifier() {
			return arc
		}
	}
	return nil
}

func (p *Net) Inputs(n Node) []*Arc {
	var inputs []*Arc
	for _, o := range p.inputs[n.Identifier()] {
		inputs = append(inputs, o)
	}
	return inputs
}

type Node interface {
	Kind() Kind
	Identifier() string
	IsNode()
}

func (p *Net) Outputs(n Node) []*Arc {
	var outputs []*Arc
	for _, o := range p.outputs[n.Identifier()] {
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
	if _, ok := p.outputs[from.Identifier()]; !ok {
		p.outputs[from.Identifier()] = make([]*Arc, 0)
	}
	p.outputs[from.Identifier()] = append(p.outputs[from.Identifier()], a)
	if _, ok := p.inputs[to.Identifier()]; !ok {
		p.inputs[to.Identifier()] = make([]*Arc, 0)
	}
	p.inputs[to.Identifier()] = append(p.inputs[to.Identifier()], a)
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
		if _, ok := net.outputs[arc.Src.Identifier()]; !ok {
			net.outputs[arc.Src.Identifier()] = make([]*Arc, 0)
		}
		net.outputs[arc.Src.Identifier()] = append(net.outputs[arc.Src.Identifier()], arc)
		if _, ok := net.inputs[arc.Dest.Identifier()]; !ok {
			net.inputs[arc.Dest.Identifier()] = make([]*Arc, 0)
		}
		net.inputs[arc.Dest.Identifier()] = append(net.inputs[arc.Dest.Identifier()], arc)
	}
	return net
}

func (p *Net) Kind() Kind { return NetObject }

type Loader[T any] interface {
	Load(io.Reader) (T, error)
}

type Flusher[T any] interface {
	Flush(io.Writer, T) error
}

type NetInput struct {
	Name        string
	Arcs        []*Arc
	Places      []*Place
	Transitions []*Transition
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

type Getter interface {
	Get(id string) (Object, error)
}

type Lister interface {
	List(f Filter) ([]Object, error)
}

type Input interface {
	IsInput()
}

type Kind int

const (
	PlaceObject Kind = iota
	TransitionObject
	ArcObject
	NetObject
)

type Object interface {
	Kind() Kind
	Identifier() string
	String() string
	Update(update Update) error
	Init(id string, input Input) error
}

type Update interface {
	IsUpdate()
}

type Filter interface {
	IsFilter()
}

type Adder interface {
	Add(Input) (Object, error)
}

type Remover interface {
	Remove(id string) (Object, error)
}

type Updater interface {
	Update(id string, update Update) (Object, error)
}

type Service interface {
	Updater
	Getter
	Lister
	Adder
	Remover
}

func (n *NetInput) IsInput()   {}
func (n *NetUpdate) IsUpdate() {}
func (n *NetFilter) IsFilter() {}
