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

type Marking map[string][]*Token

func (m Marking) Copy() Marking {
	ret := make(Marking)
	for k, v := range m {
		ret[k] = make([]*Token, len(v))
		for i, t := range v {
			ret[k][i] = t.Copy()
		}
	}
	return ret
}

func (m Marking) PlaceTokens(place *Place, tokens ...*Token) error {
	if _, ok := m[place.Identifier()]; !ok {
		return errors.New("place not found")
	}
	for _, t := range tokens {
		if !place.CanAccept(t.Schema) {
			return errors.New("token not accepted")
		}
		if len(m[place.Identifier()]) >= place.Bound {
			return errors.New("place is full")
		}
		m[place.Identifier()] = append(m[place.Identifier()], t)
	}
	return nil
}

// Net struct
type Net struct {
	Name        string
	Places      []*Place
	Transitions []*Transition
	Arcs        []*Arc
	inputs      map[string][]*Arc
	outputs     map[string][]*Arc
}

func (p *Net) Document() Document {
	//TODO implement me
	panic("implement me")
}

func (p *Net) From(doc Document) error {
	//TODO implement me
	panic("implement me")
}

func (p *Net) NewMarking() Marking {
	m := make(Marking)
	for _, place := range p.Places {
		m[place.Identifier()] = make([]*Token, 0)
	}
	return m
}

func (p *Net) Init(input Input) error {
	in, ok := input.(*NetInput)
	if !ok {
		return ErrWrongInput
	}
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

func (p *Net) Enabled(marking Marking, t *Transition) bool {
	for _, arc := range p.Inputs(t) {
		if pt, ok := arc.Src.(*Place); ok {
			if len(marking[pt.Identifier()]) == 0 {
				return false
			}
		}
	}
	return true
}

func (p *Net) EnabledTransitions(m Marking) []*Transition {
	var transitions []*Transition
	for _, t := range p.Transitions {
		if p.Enabled(m, t) {
			transitions = append(transitions, t)
		}
	}
	return transitions
}

func (p *Net) Process(m Marking) (Marking, error) {
	t := p.EnabledTransitions(m)[0]
	return p.Fire(m, t)
}

func (p *Net) Fire(m Marking, t *Transition) (Marking, error) {
	var tok *Token
	ret := m.Copy()
	for _, arc := range p.Inputs(t) {
		if pt, ok := arc.Src.(*Place); ok {
			// pop a token from the place
			tok = m[pt.Identifier()][0]
			ret[pt.Identifier()] = m[pt.Identifier()][1:]
		}
	}

	if tok == nil {
		return m, errors.New("no token found")
	}

	out, err := t.Handle(tok)
	if err != nil {
		return m, err
	}

	for _, arc := range p.Outputs(t) {
		if pt, ok := arc.Dest.(*Place); ok {
			if len(ret[pt.Identifier()]) >= pt.Bound {
				return m, errors.New("place is full")
			}
			ret[pt.Identifier()] = append(ret[pt.Identifier()], out)
		}
	}
	return ret, nil
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
	return p.Name
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

func NewNet(name string) *Net {
	return &Net{
		Name:        name,
		Places:      make([]*Place, 0),
		Transitions: make([]*Transition, 0),
		Arcs:        make([]*Arc, 0),
		inputs:      make(map[string][]*Arc),
		outputs:     make(map[string][]*Arc),
	}
}

func (p *Net) WithPlaces(places ...*Place) *Net {
	p.Places = append(p.Places, places...)
	return p
}

func (p *Net) WithTransitions(transitions ...*Transition) *Net {
	p.Transitions = append(p.Transitions, transitions...)
	return p
}

func (p *Net) WithArcs(arcs ...*Arc) *Net {
	for _, arc := range arcs {
		_, err := p.AddArc(arc.Src, arc.Dest)
		if err != nil {
			panic(err)
		}
	}
	return p
}

func LoadNet(places []*Place, transitions []*Transition, arcs []*Arc) *Net {

	for _, p := range places {
		if p.Bound == 0 {
			p.Bound = 1
		}
	}
	net := &Net{
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

func (n *NetInput) Object() Object {
	//TODO implement me
	panic("implement me")
}

func (n *NetInput) Kind() Kind {
	return NetObject
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

func (n *NetFilter) Filter() Document {
	//TODO implement me
	panic("implement me")
}

func (n *NetInput) IsInput()   {}
func (n *NetUpdate) IsUpdate() {}
func (n *NetFilter) IsFilter() {}
