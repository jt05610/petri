package petri

import (
	"context"
	"errors"
	"io"
)

var ErrWrongInput = errors.New("wrong input")
var ErrWrongUpdate = errors.New("wrong update")

var _ Object = (*Net)(nil)
var _ Input = (*NetInput)(nil)
var _ Update = (*NetUpdate)(nil)
var _ Filter = (*NetFilter)(nil)

type Marking map[string][]*Token[interface{}]

func (m Marking) Copy() Marking {
	ret := make(Marking)
	for k, v := range m {
		ret[k] = make([]*Token[interface{}], len(v))
		for i, t := range v {
			ret[k][i] = t
		}
	}
	return ret
}

func (m Marking) TokenMap(place *Place) map[string]*Token[interface{}] {
	tokens := m[place.String()]
	tokMap := make(map[string]*Token[interface{}])
	for _, t := range tokens {
		if _, ok := tokMap[t.Schema.Name]; ok {
			continue
		}
		tokMap[t.Schema.Name] = t
	}
	return tokMap
}

func (m Marking) Get(place *Place, schema string) *Token[interface{}] {
	if _, ok := m[place.String()]; !ok {
		return nil
	}
	for _, t := range m[place.String()] {
		if t.Schema.Name == schema {
			return t
		}
	}
	return nil
}

func (m Marking) Remove(place *Place, token *Token[interface{}]) {
	if _, ok := m[place.String()]; !ok {
		return
	}
	for i, t := range m[place.String()] {
		if t.Schema.Name == token.Schema.Name {
			m[place.String()] = append(m[place.String()][:i], m[place.String()][i+1:]...)
			return
		}
	}
}

func (m Marking) PlaceTokens(place *Place, tokens ...*Token[interface{}]) error {
	if _, ok := m[place.String()]; !ok {
		return errors.New("place not found")
	}
	for _, t := range tokens {
		if !place.CanAccept(t.Schema) {
			return errors.New("token not accepted")
		}
		if len(m[place.String()]) >= place.Bound {
			return errors.New("place is full")
		}
		m[place.String()] = append(m[place.String()], t)
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
		m[place.String()] = make([]*Token[interface{}], 0)
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
			if len(marking[pt.String()]) == 0 {
				return false
			}
		}
	}
	return true
}

func (p *Net) EnabledTransitions(m Marking, events ...string) []*Transition {
	var transitions []*Transition
	for _, t := range p.Transitions {
		if p.Enabled(m, t) {
			if t.Cold {
				for _, e := range events {
					if t.Name == e {
						transitions = append(transitions, t)
					}
				}
				continue
			}
			transitions = append(transitions, t)
		}
	}
	return transitions
}

type Event[T any] struct {
	Name string `json:"name"`
	Data T      `json:"data"`
}

func (p *Net) Process(m Marking, events ...Event[any]) (Marking, error) {
	eventNames := make([]string, 0)
	for _, e := range events {
		eventNames = append(eventNames, e.Name)
	}

	for _, t := range p.EnabledTransitions(m, eventNames...) {
		m, err := p.Fire(m, t, events...)
		if err != nil {
			continue
		}
		return p.Process(m)
	}

	return m, nil

}

func IndexTokenByType(tokens []*Token[interface{}]) map[string]*Token[interface{}] {
	index := make(map[string]*Token[interface{}])
	for _, t := range tokens {
		if _, ok := index[t.Schema.Name]; ok {
			continue
		}
		index[t.Schema.Name] = t
	}
	return index
}

func (p *Net) Fire(m Marking, t *Transition, events ...Event[any]) (Marking, error) {
	tokens := make([]*Token[interface{}], 0)
	handlerMap := make(map[string]EventFunc[any, any])
	dataMap := make(map[string]interface{})
	for _, e := range events {
		handlerMap[e.Name] = p.RouteEvent(e)
		dataMap[e.Name] = e.Data
	}
	handler, hasHandler := handlerMap[t.Name]
	data := dataMap[t.Name]

	ret := m.Copy()

	for _, arc := range p.Inputs(t) {
		if pt, ok := arc.Src.(*Place); ok {
			tok, err := arc.TakeToken(ret)
			if err != nil {
				return m, err
			}
			tokens = append(tokens, tok)
			ret.Remove(pt, tok)
		}
	}
	if len(tokens) == 0 && !hasHandler {
		return m, errors.New("no tokens found")
	}

	var err error

	if !t.CanFire(IndexTokenByType(tokens)) {
		return m, errors.New("transition cannot fire")
	}

	if t.Handler != nil {
		tokens, err = t.Handle(tokens...)
		if err != nil {
			return m, err
		}
	}

	hasOutputs := len(p.Outputs(t)) > 0

	if hasHandler {
		var eventResult interface{}
		if hasOutputs {
			eventResult, err = handler(context.Background(), data)
		} else {
			t := tokens[0]
			tokData := t.Value
			_, err = handler(context.Background(), tokData)
		}
		if err != nil {
			return m, err
		}
		for _, arc := range p.Outputs(t) {
			if _, ok := arc.Dest.(*Place); ok {
				tok, err := arc.OutputSchema.NewToken(eventResult)
				if err != nil {
					return m, err
				}
				tokens = append(tokens, tok)
			}
		}
	}

	if len(tokens) == 0 {
		return m, errors.New("no tokens found")
	}

	tokenIndex := IndexTokenByType(tokens)

	for _, arc := range p.Outputs(t) {
		if _, ok := arc.Dest.(*Place); ok {
			err := arc.PlaceToken(ret, tokenIndex)
			if err != nil {
				return m, err
			}
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

func (p *Net) AddArc(arc *Arc) error {
	if arc.Src.Kind() == arc.Dest.Kind() {
		return errors.New("cannot connect two places or two transitions")
	}
	if arc := p.Arc(arc.Src, arc.Dest); arc != nil {
		return errors.New("arc already exists")
	}

	p.Arcs = append(p.Arcs, arc)
	if _, ok := p.outputs[arc.Src.Identifier()]; !ok {
		p.outputs[arc.Src.Identifier()] = make([]*Arc, 0)
	}
	p.outputs[arc.Src.Identifier()] = append(p.outputs[arc.Src.Identifier()], arc)
	if _, ok := p.inputs[arc.Dest.Identifier()]; !ok {
		p.inputs[arc.Dest.Identifier()] = make([]*Arc, 0)
	}
	p.inputs[arc.Dest.Identifier()] = append(p.inputs[arc.Dest.Identifier()], arc)
	return nil
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
		err := p.AddArc(arc)
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

func (p *Net) RouteEvent(event Event[any]) EventFunc[any, any] {
	for _, t := range p.Transitions {
		if t.Name == event.Name {
			return t.Event
		}
	}
	return nil
}

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
