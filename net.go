package petri

import (
	"context"
	"errors"
	"io"
	"strings"
)

var ErrWrongInput = errors.New("wrong input")
var ErrWrongUpdate = errors.New("wrong update")

type MarkingService map[string]TokenQueue

type Marking map[string][]Token

func (m MarkingService) Equals(other MarkingService) bool {
	if len(m) != len(other) {
		return false
	}
	for k, v := range m {
		if otherV, ok := other[k]; !ok {
			return false
		} else {
			n, err := v.Available(context.Background())
			if err != nil {
				return false
			}
			otherN, err := otherV.Available(context.Background())
			if err != nil {
				return false
			}
			if n != otherN {
				return false
			}
		}
	}
	return true
}

func (m MarkingService) Mark() Marking {
	ret := make(Marking)
	for k, v := range m {
		ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
		defer can()
		t, err := v.Peek(ctx)
		if err != nil {
			panic(err)
		}
		ret[k] = t
	}
	return ret
}

func (m MarkingService) Copy() MarkingService {
	ret := make(MarkingService)
	for k, v := range m {
		ret[k] = v.Copy()
	}
	return ret
}

func (m MarkingService) TokenMap(place *Place) map[string]Token {
	tokens := m[place.ID]
	tokMap := make(map[string]Token)
	ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
	defer can()
	tok, err := tokens.Peek(ctx)
	if err != nil {
		panic(err)
	}
	tokMap[place.AcceptedTokens[0].Name] = tok[0]
	return tokMap
}

func (m MarkingService) Get(place *Place, schema string) Token {
	if pl, ok := m[place.String()]; !ok {
		return Token{}
	} else {
		ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
		defer can()
		t, err := pl.Dequeue(ctx)
		if err != nil {
			panic(err)
		}
		return t
	}
}

func (m MarkingService) PlaceTokens(place *Place, tokens ...Token) error {
	if pl, ok := m[place.ID]; !ok {
		return errors.New("place not found")
	} else {
		ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
		defer can()
		return pl.Enqueue(ctx, tokens...)
	}
}

// Net struct
type Net struct {
	ID           string                  `json:"_id"`
	Name         string                  `json:"name"`
	TokenSchemas map[string]*TokenSchema `json:"tokenSchemas,omitempty"`
	Places       map[string]*Place       `json:"places,omitempty"`
	Transitions  map[string]*Transition  `json:"transitions,omitempty"`
	Arcs         []*Arc                  `json:"arcs,omitempty"`
	Nets         []*Net                  `json:"nets,omitempty"`
	inputs       map[string][]*Arc
	outputs      map[string][]*Arc
}

// Local returns the local places of the net that do not belong to any subnets
func (p *Net) Local() map[string]*Place {
	local := make(map[string]*Place)
	for k, pl := range p.Places {
		if !strings.Contains(k, ".") {
			local[k] = pl
		}
	}
	return local
}

func (p *Net) Subnet(name string) *Net {
	for _, net := range p.Nets {
		if net.Name == name {
			return net
		}
	}
	return nil
}

func (p *Net) SubnetArcs(name string) []*Arc {
	arcs := make([]*Arc, 0)
	if net := p.Subnet(name); net == nil {
		return nil
	}
	for _, arc := range p.Arcs {
		if strings.Contains(arc.Src.Identifier(), name+".") || strings.Contains(arc.Dest.Identifier(), name+".") {
			arcs = append(arcs, arc)
		}
	}
	return arcs
}

func (p *Net) Parent(n Node) *Net {
	if !strings.Contains(n.Identifier(), ".") {
		return p
	}
	parent := strings.Split(n.Identifier(), ".")[0]
	for _, net := range p.Nets {
		if net.Name == parent {
			return net
		}
	}
	return nil
}

func (p *Net) Owns(t *Transition) bool {
	if _, ok := p.Transitions[t.Name]; ok {
		return true
	}
	return false
}

func (p *Net) InputSchema(n string) *TokenSchema {
	arc, found := p.inputs[n]
	if !found {
		return nil
	}
	return arc[0].OutputSchema
}

func (p *Net) OutputSchema(n string) *TokenSchema {
	arc, found := p.outputs[n]
	if !found {
		return nil
	}
	return arc[0].OutputSchema
}

func (p *Net) makeInputsOutputs() {
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

}
func (p *Net) PostInit() error {
	return nil
}

func (p *Net) Document() Document {
	return Document{
		"_id":          p.ID,
		"name":         p.Name,
		"tokenSchemas": p.TokenSchemas,
		"places":       p.Places,
		"transitions":  p.Transitions,
		"arcs":         p.Arcs,
		"nets":         p.Nets,
	}
}

func (p *Net) WithoutArc(a *Arc) *Net {
	for i, arc := range p.Arcs {
		if arc == a {
			p.Arcs = append(p.Arcs[:i], p.Arcs[i+1:]...)
			break
		}
	}
	return p
}

func (p *Net) WithoutPlace(place *Place) *Net {
	if _, ok := p.Places[place.ID]; !ok {
		return p
	}
	delete(p.Places, place.ID)
	return p
}

func (p *Net) WithoutTransition(t *Transition) *Net {
	if _, ok := p.Transitions[t.ID]; !ok {
		return p
	}
	delete(p.Transitions, t.ID)
	return p
}

func (p *Net) NewMarking() MarkingService {
	m := make(MarkingService)
	for _, pl := range p.Places {
		m[pl.ID] = pl.TokenQueue
	}
	return m
}

func (p *Net) InPlaces() []*Place {
	var places []*Place
	for _, pl := range p.Places {
		if len(p.inputs[pl.ID]) == 0 {
			places = append(places, pl)
		}
	}
	return places
}

func (p *Net) OutPlaces() []*Place {
	var places []*Place
	for _, pl := range p.Places {
		if len(p.outputs[pl.ID]) == 0 {
			places = append(places, pl)
		}
	}
	return places
}

func (p *Net) Place(name string) *Place {
	return p.Places[name]
}

func (p *Net) Transition(name string) *Transition {
	return p.Transitions[name]
}

func (p *Net) Init(input Input) error {
	in, ok := input.(*NetInput)
	if !ok {
		return ErrWrongInput
	}
	p.Name = in.Name
	p.Places = CreateIndex(in.Places)
	p.Transitions = CreateIndex(in.Transitions)
	p.Arcs = in.Arcs
	p.TokenSchemas = CreateIndex(in.TokenSchemas)
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
			tokens := marking[pt.ID]
			if len(tokens) == 0 {
				return false
			}
			if !pt.CanAccept(tokens[0].Schema) {
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

type Event[T any] struct {
	Name string `json:"name"`
	Data T      `json:"data"`
}

var (
	ErrNoEvents = errors.New("no events enabled")
)

func (p *Net) Process(m Marking) (Marking, error) {
	enabled := p.EnabledTransitions(m)
	if len(enabled) == 0 {
		return m, ErrNoEvents
	}
	m, err := p.Fire(m, enabled[0])
	if err != nil {
		return m, err
	}

	processed, err := p.Process(m)
	if err != nil {
		if errors.Is(err, ErrNoEvents) {
			return processed, nil
		}
		return m, err
	}
	return processed, nil
}

func IndexTokenByType(tokens []Token) map[string]Token {
	index := make(map[string]Token)
	for _, t := range tokens {
		if _, ok := index[t.Schema.Name]; ok {
			continue
		}
		index[t.Schema.Name] = t
	}
	return index
}

func (p *Net) Fire(m Marking, t *Transition) (Marking, error) {
	tokens := make([]Token, 0)

	for _, arc := range p.Inputs(t) {
		if _, ok := arc.Src.(*Place); ok {
			pl := arc.Place
			tok, err := arc.TakeToken()
			if err != nil {
				return m, err
			}
			tokens = append(tokens, tok)
			// remove token from marking
			m[pl.ID] = m[pl.ID][1:]
		}
	}
	if len(tokens) == 0 {
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

	if len(tokens) == 0 {
		return m, errors.New("no tokens found")
	}

	tokenIndex := IndexTokenByType(tokens)

	for _, arc := range p.Outputs(t) {
		if _, ok := arc.Dest.(*Place); ok {
			err := arc.PlaceToken(tokenIndex)
			if err != nil {
				return m, err
			}
			m[arc.Dest.Identifier()] = append(m[arc.Dest.Identifier()], tokens[0])
		}
	}
	return m, nil
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
		p.Places = CreateIndex(u.Input.Places)
	}
	if u.Mask.Transitions {
		p.Transitions = CreateIndex(u.Input.Transitions)
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
		ID:           ID(),
		Name:         name,
		Places:       make(map[string]*Place),
		Transitions:  make(map[string]*Transition),
		Arcs:         make([]*Arc, 0),
		Nets:         make([]*Net, 0),
		TokenSchemas: make(map[string]*TokenSchema),
		inputs:       make(map[string][]*Arc),
		outputs:      make(map[string][]*Arc),
	}
}

func (p *Net) WithPlaces(places ...*Place) *Net {
	for _, pl := range places {
		if pl.Bound == 0 {
			pl.Bound = 1
		}
		for _, tok := range pl.AcceptedTokens {
			p.TokenSchemas[tok.Name] = tok
		}
		p.Places[pl.Name] = pl
	}
	return p
}

func (p *Net) ReplacePlace(old, new *Place) *Net {
	if _, ok := p.Places[old.Name]; !ok {
		return p
	}
	p.Places[new.Name] = new
	for _, arc := range p.Inputs(old) {
		arc.Dest = new
		arc.Place = new
		p.inputs[new.Identifier()] = append(p.inputs[new.Identifier()], arc)
	}
	delete(p.inputs, old.Name)
	for _, arc := range p.Outputs(old) {
		arc.Src = new
		arc.Place = new
		p.outputs[new.Identifier()] = append(p.outputs[new.Identifier()], arc)
	}
	delete(p.outputs, old.Name)
	delete(p.Places, old.Name)
	return p
}

func (p *Net) JoinPlaces(p1 *Place, p2 *Place) *Net {
	iterPlace := NewPlace(p1.Name, p1.Bound, p1.AcceptedTokens...)
	net := p.WithPlaces(iterPlace)
	parent1 := p.Parent(p1)
	parent1 = parent1.ReplacePlace(p1, iterPlace)
	parent2 := p.Parent(p2)
	parent2 = parent2.ReplacePlace(p2, iterPlace)
	delete(p.Places, p1.ID)
	delete(p.Places, p2.ID)
	p.inputs[iterPlace.Identifier()] = append(p.inputs[iterPlace.Identifier()], p.inputs[p1.Identifier()]...)
	p.inputs[iterPlace.Identifier()] = append(p.inputs[iterPlace.Identifier()], p.inputs[p2.Identifier()]...)
	p.outputs[iterPlace.Identifier()] = append(p.outputs[iterPlace.Identifier()], p.outputs[p1.Identifier()]...)
	p.outputs[iterPlace.Identifier()] = append(p.outputs[iterPlace.Identifier()], p.outputs[p2.Identifier()]...)
	delete(p.inputs, p1.Identifier())
	delete(p.inputs, p2.Identifier())

	return net
}

func (p *Net) InputPlaces() []*Place {
	var places []*Place
	for _, pl := range p.Places {
		if len(p.inputs[pl.ID]) == 0 {
			places = append(places, pl)
		}
	}
	return places
}

func (p *Net) OutputPlaces() []*Place {
	var places []*Place
	for _, pl := range p.Places {
		if len(p.outputs[pl.ID]) == 0 {
			places = append(places, pl)
		}
	}
	return places
}

func (p *Net) WithTransitions(transitions ...*Transition) *Net {
	for _, t := range transitions {
		p.Transitions[t.Name] = t
	}
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

func makeName(ss ...string) string {
	return strings.Join(ss, ".")
}

func (p *Net) AddSubNet(n *Net) {
	p.Nets = append(p.Nets, n)
	netName := n.Name
	idMap := make(map[string]string)
	for pn, pl := range n.Places {
		idMap[pl.ID] = makeName(netName, pn)
		pl.ID = makeName(netName, pn)
		p.Places[makeName(netName, pn)] = pl
	}
	for pn, tr := range n.Transitions {
		idMap[tr.ID] = makeName(netName, pn)
		tr.ID = makeName(netName, pn)
		p.Transitions[makeName(netName, pn)] = tr
	}
	for oldId, newId := range idMap {
		n.inputs[newId] = n.inputs[oldId]
		n.outputs[newId] = n.outputs[oldId]
		delete(n.inputs, oldId)
		delete(n.outputs, oldId)
	}

	for _, a := range n.Arcs {
		err := p.AddArc(a)
		if err != nil {
			panic(err)
		}
	}
}

func (p *Net) WithNets(nets ...*Net) *Net {
	for _, n := range nets {
		p.AddSubNet(n)
	}
	return p
}

func (p *Net) WithTokenSchemas(schemas ...*TokenSchema) *Net {
	for _, schema := range schemas {
		p.TokenSchemas[schema.Name] = schema
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
		Places:       CreateIndex(places),
		Transitions:  CreateIndex(transitions),
		Arcs:         arcs,
		TokenSchemas: make(map[string]*TokenSchema),
		Nets:         make([]*Net, 0),
		inputs:       make(map[string][]*Arc),
		outputs:      make(map[string][]*Arc),
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

func (p *Net) Node(k string) Node {
	if n := p.Place(k); n != nil {
		return n
	}
	if t := p.Transition(k); t != nil {
		return t
	}
	return nil
}

type Loader[T any] interface {
	Load(io.Reader) (T, error)
}

type Flusher[T any] interface {
	Flush(io.Writer, T) error
}

func CreateIndex[T Indexable](t []T) map[string]T {
	index := make(map[string]T)
	for _, v := range t {
		index[v.Index()] = v
	}
	return index
}

type NetInput struct {
	Name         string         `json:"name"`
	TokenSchemas []*TokenSchema `json:"tokenSchemas,omitempty"`
	Arcs         []*Arc         `json:"arcs,omitempty"`
	Places       []*Place       `json:"places,omitempty"`
	Transitions  []*Transition  `json:"transitions,omitempty"`
	Nets         []*Net         `json:"nets,omitempty"`
}

func (n *NetInput) Object() Object {
	return NewNet(n.Name).WithPlaces(n.Places...).WithTransitions(n.Transitions...).WithArcs(n.Arcs...).WithNets(n.Nets...).WithTokenSchemas(n.TokenSchemas...)
}

func (n *NetInput) Kind() Kind {
	return NetObject
}

type NetMask struct {
	TokenSchemas bool `json:"tokenSchemas,omitempty"`
	Name         bool `json:"name,omitempty"`
	Places       bool `json:"places,omitempty"`
	Transitions  bool `json:"transitions,omitempty"`
	Arcs         bool `json:"arcs,omitempty"`
	Nets         bool `json:"nets,omitempty"`
}

type NetUpdate struct {
	Input *NetInput
	Mask  *NetMask
}

type NetFilter struct {
	ID   *StringSelector `json:"_id,omitempty"`
	Name *StringSelector `json:"name,omitempty"`
}

func (n *NetFilter) Filter() Document {
	//TODO implement me
	panic("implement me")
}

func (n *NetInput) IsInput()   {}
func (n *NetUpdate) IsUpdate() {}
func (n *NetFilter) IsFilter() {}
