package petrifile

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile"
	"strings"
)

type Place interface {
	Place(name string, tokenMap map[string]*petri.TokenSchema) *petri.Place
}

type StringPlace string
type ListPlace []string
type DefinedPlace struct {
	Accepts string
	Bound   int
}

var (
	_ Place = (*StringPlace)(nil)
	_ Place = (*DefinedPlace)(nil)
)

func (s StringPlace) Place(name string, tokenMap map[string]*petri.TokenSchema) *petri.Place {
	accept, ok := tokenMap[string(s)]
	if !ok {
		panic(fmt.Sprintf("unknown token type %s", s))
	}
	return petri.NewPlace(name, 1, accept)
}

func (s DefinedPlace) Place(name string, tokenMap map[string]*petri.TokenSchema) *petri.Place {
	accept, ok := tokenMap[s.Accepts]
	if !ok {
		panic(fmt.Sprintf("unknown token type %s", s.Accepts))
	}
	return petri.NewPlace(name, s.Bound, accept)
}

type Arc interface {
	ToPlace(net *petri.Net, from *petri.Transition) *petri.Arc
	FromPlace(net *petri.Net, to *petri.Transition) *petri.Arc
}

type StringArc string

func (s StringArc) FromPlace(net *petri.Net, to *petri.Transition) *petri.Arc {
	pl := net.Place(string(s))
	if pl == nil {
		panic(fmt.Sprintf("unknown place %s", s))
	}
	return petri.NewArc(pl, to, pl.Schema.Name, pl.Schema)
}

func (s StringArc) ToPlace(net *petri.Net, from *petri.Transition) *petri.Arc {
	pl := net.Place(string(s))
	if pl == nil {
		panic(fmt.Sprintf("unknown place %s", s))
	}
	return petri.NewArc(from, pl, pl.Schema.Name, pl.Schema)
}

type ExpressionArc map[string]interface{}

func ToMapString(i interface{}) Expr {
	switch v := i.(type) {
	case string:
		return Expr{
			Op:     v,
			Fields: nil,
		}
	case map[string]interface{}:
		ret := make(map[string]Expr)
		fields := make([]string, 0, len(v))
		for k, v := range v {
			ret[k] = ToMapString(v)
			fields = append(fields, k)
		}
		bld := strings.Builder{}
		bld.WriteString("{")
		for k, v := range ret {
			bld.WriteString(fmt.Sprintf("\"%s\": %s, ", k, v))
		}
		bld.WriteString("}")
		return Expr{
			Op:     bld.String(),
			Fields: fields,
		}
	}
	panic(fmt.Sprintf("unknown input type %T", i))
}

type Expr struct {
	Op     string
	Fields []string
}

func (e Expr) String() string {
	return e.Op
}

func (e ExpressionArc) ToTokExprMap() map[string]Expr {
	ret := make(map[string]Expr)
	for k, v := range e {
		ret[k] = ToMapString(v)
	}
	return ret
}

func (e ExpressionArc) ToPlace(net *petri.Net, from *petri.Transition) *petri.Arc {
	for k, v := range e.ToTokExprMap() {
		pl := net.Place(k)
		if pl == nil {
			panic(fmt.Sprintf("unknown place %s", k))
		}
		return petri.NewArc(from, pl, v.Op, pl.Schema)
	}
	panic(fmt.Sprintf("unknown token type %s", e))
}

func (e ExpressionArc) FromPlace(net *petri.Net, to *petri.Transition) *petri.Arc {
	for k, v := range e.ToTokExprMap() {
		pl := net.Place(k)
		if pl == nil {
			panic(fmt.Sprintf("unknown place %s", k))
		}
		return petri.NewArc(pl, to, v.Op, pl.Schema)
	}
	panic(fmt.Sprintf("unknown token type %s", e))
}

var (
	_ Arc = (*StringArc)(nil)
	_ Arc = (*ExpressionArc)(nil)
)

type Transition struct {
	Event   bool `yaml:"event,omitempty"`
	Outputs interface{}
	Inputs  interface{}
	Guard   []string
}

type PlaceRef interface {
	Arcs(tok string) ([]*petri.Arc, error)
}

type Link struct {
	From interface{}
	To   interface{}
	net  *petri.Net
}

type PlaceArg struct {
	petri.Node
	*petri.TokenSchema
	Expr
}

func (l *Link) places(s interface{}) ([]*PlaceArg, error) {
	ret := make([]*PlaceArg, 0)
	switch s := s.(type) {
	case string:
		pl := l.net.Node(s)
		if pl == nil {
			return nil, fmt.Errorf("unknown place %s", s)
		}
		ret = append(ret, &PlaceArg{
			Node: pl,
		})
	case []interface{}:
		for _, v := range s {
			switch v := v.(type) {
			case string:
				pl := l.net.Node(v)
				if pl == nil {
					return nil, fmt.Errorf("unknown place %s", v)
				}
				ret = append(ret, &PlaceArg{
					Node: pl,
				})
			case map[string]interface{}:
				for k, v := range v {
					pl := l.net.Node(k)
					if pl == nil {
						return nil, fmt.Errorf("unknown place %s", k)
					}
					ret = append(ret, &PlaceArg{
						Node: pl,
						Expr: ToMapString(v),
					})
				}
			}
		}
	case map[string]interface{}:
		for k, v := range s {
			pl := l.net.Node(k)
			if pl == nil {
				return nil, fmt.Errorf("unknown place %s", k)
			}
			ret = append(ret, &PlaceArg{
				Node: pl,
				Expr: ToMapString(v),
			})
		}
	}
	return ret, nil
}

func (l *Link) Arcs() ([]*petri.Arc, error) {
	aa := make([]*petri.Arc, 0)
	from, err := l.places(l.From)
	if err != nil {
		return nil, err
	}
	to, err := l.places(l.To)
	if err != nil {
		return nil, err
	}
	for _, f := range from {
		for _, t := range to {
			if t.TokenSchema == nil {
				pl := l.net.Place(f.Node.Identifier())
				if pl == nil {
					pl = l.net.Place(t.Node.Identifier())
				}
				t.TokenSchema = pl.Schema
			}
			if t.Op == "" {
				t.Op = t.TokenSchema.Name
			}
			arc := petri.NewArc(f.Node, t.Node, t.Op, t.TokenSchema)
			arc.LinksNets = true
			arc.PlaceNet = l.net.Parent(arc.Place)
			arc.TransitionNet = l.net.Parent(arc.Transition)
			aa = append(aa, arc)
		}
	}
	return aa, nil
}

type Petrifile struct {
	Petri            petrifile.Version
	Version          string
	Name             string
	Imports          []string
	Types            map[string]petri.Type
	Places           map[string]interface{}
	Transitions      map[string]Transition
	Nets             map[string]string
	Joins            []string
	Links            []Link
	net              *petri.Net
	arcs             []*petri.Arc
	*builder.Builder `yaml:"-"`
}

func ParseInput(i interface{}) Arc {
	switch v := i.(type) {
	case string:
		return StringArc(v)
	case map[string]interface{}:
		return ExpressionArc(v)
	}
	panic(fmt.Sprintf("unknown input type %T", i))
}

func ParseOutput(i interface{}) Arc {
	switch v := i.(type) {
	case string:
		return StringArc(v)
	case map[string]interface{}:
		return ExpressionArc(v)
	}
	panic(fmt.Sprintf("unknown output type %T", i))
}

func PrimitiveTokenMap() map[string]*petri.TokenSchema {
	return map[string]*petri.TokenSchema{
		"string": petri.String(),
		"int":    petri.Integer(),
		"float":  petri.Float64(),
		"bool":   petri.Boolean(),
		"signal": petri.Signal(),
		"time":   petri.Time(),
	}
}

func (p *Petrifile) makePlaces() []*petri.Place {
	ret := make([]*petri.Place, 0, len(p.Places))
	tm := PrimitiveTokenMap()
	for k, v := range p.net.TokenSchemas {
		tm[k] = v
	}

	for n, v := range p.Places {
		switch v := v.(type) {
		case string:
			ret = append(ret, StringPlace(v).Place(n, tm))
		case StringPlace:
			ret = append(ret, v.Place(n, tm))
		case DefinedPlace:
			ret = append(ret, v.Place(n, tm))
		}
	}
	return ret
}

func (p *Petrifile) makeTransitions() []*petri.Transition {
	ret := make([]*petri.Transition, 0, len(p.Transitions))
	arcs := make([]*petri.Arc, 0)
	for n, v := range p.Transitions {
		t := petri.NewTransition(n)
		if v.Guard != nil {
			exp := strings.Join(v.Guard, " && ")
			t.Expression = exp
		}

		ret = append(ret, t)
		var schema *petri.TokenSchema
		if v.Inputs != nil {
			switch in := v.Inputs.(type) {
			case []interface{}:
				for _, v := range in {
					a := ParseInput(v).FromPlace(p.net, t)
					arcs = append(arcs, a)
				}
			case interface{}:
				a := ParseInput(in).FromPlace(p.net, t)
				arcs = append(arcs, a)
				schema = a.Src.(*petri.Place).Schema
			}
		}

		if v.Outputs != nil {
			switch out := v.Outputs.(type) {
			case nil:
				fmt.Printf("no outputs for transition %s\n", n)
			case []interface{}:
				for _, v := range out {
					a := ParseOutput(v).ToPlace(p.net, t)
					arcs = append(arcs, a)
				}
			case interface{}:
				a := ParseOutput(out).ToPlace(p.net, t)
				arcs = append(arcs, a)
				schema = a.Dest.(*petri.Place).Schema
			}
		}
		if v.Event {
			t = t.WithEvent(schema)
		}
	}
	p.arcs = arcs
	return ret
}

func (p *Petrifile) makeArcs() []*petri.Arc {
	return p.arcs
}

func (p *Petrifile) makeJoins() {
	if p.Joins == nil {
		return
	}
	for _, v := range p.Joins {
		places := strings.Split(v, "->")
		if len(places) != 2 {
			panic(fmt.Sprintf("join %s must have 2 places", v))
		}
		pl1 := p.net.Place(places[0])
		pl2 := p.net.Place(places[1])
		if pl1 == nil {
			panic(fmt.Sprintf("unknown place %s", places[0]))
		}
		if pl2 == nil {
			panic(fmt.Sprintf("unknown place %s", places[1]))
		}
		p.net = p.net.JoinPlaces(pl1, pl2)
	}
}

func (p *Petrifile) importTypes(fName string) map[string]petri.Type {
	n, err := p.Builder.Build(context.Background(), fName)
	if err != nil {
		panic(err)
	}
	if n == nil {
		return nil
	}
	return petri.NetTypes(n)
}

func (p *Petrifile) Net() *petri.Net {
	p.net = petri.NewNet(p.Name)
	types := make(map[string]petri.Type)
	for k, v := range p.Types {
		types[k] = v
	}
	for _, v := range p.Imports {
		ty := p.importTypes(v + ".yaml")
		for k, ty := range ty {
			types[fmt.Sprintf("%s.%s", v, k)] = ty
		}
	}
	typeGraph := petri.BuildTypeGraph(types)
	p.net = p.net.
		WithTokenSchemas(typeGraph.Types()...).
		WithPlaces(p.makePlaces()...).
		WithTransitions(p.makeTransitions()...).
		WithNets(p.makeSubNets()...).
		WithArcs(p.makeArcs()...)

	if p.Nets != nil && p.Links != nil {
		for _, v := range p.Links {
			v.net = p.net
			aa, err := v.Arcs()
			if err != nil {
				panic(err)
			}
			p.net = p.net.WithArcs(aa...)
		}
	}
	p.makeJoins()
	return p.net
}

func (p *Petrifile) makeSubNets() []*petri.Net {
	nn := make([]*petri.Net, 0, len(p.Nets))
	for k, v := range p.Nets {
		if k == p.Name {
			panic(fmt.Sprintf("net %s cannot be a sub net of itself", k))
		}
		n, err := p.Builder.Build(context.Background(), v)
		if err != nil {
			panic(err)
		}
		nn = append(nn, n)
	}
	for _, imp := range p.Imports {
		n, err := p.Builder.Build(context.Background(), imp+".yaml")
		if err != nil {
			panic(err)
		}
		nn = append(nn, n)
	}

	return nn
}

func New(n *petri.Net) *Petrifile {
	return &Petrifile{
		net: n,
	}
}
