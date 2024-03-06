package petrifile

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/petrifile"
	"strings"
)

type Type map[string]string

func (t Type) Schema(name string, lookup map[string]map[string]petri.Properties) *petri.TokenSchema {
	s := petri.NewTokenSchema(name)
	props := make(map[string]petri.Properties)
	for k, v := range t {
		if petri.TokenType(v).IsPrimitive() {
			props[k] = petri.Properties{
				Type:       petri.TokenType(v),
				Properties: nil,
			}
			continue
		}
		p, ok := lookup[v]
		if !ok {
			panic(fmt.Sprintf("unknown type %s. Please declare it", v))
		}
		props[k] = petri.Properties{
			Type:       petri.Obj,
			Properties: p,
		}
	}
	lookup[name] = props
	return s.WithProperties(props)
}

func (t Type) IsNested() bool {
	for _, v := range t {
		if !petri.TokenType(v).IsPrimitive() {
			return true
		}
	}
	return false
}

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
	_ Place = (*ListPlace)(nil)
	_ Place = (*DefinedPlace)(nil)
)

func (s StringPlace) Place(name string, tokenMap map[string]*petri.TokenSchema) *petri.Place {
	accept, ok := tokenMap[string(s)]
	if !ok {
		panic(fmt.Sprintf("unknown token type %s", s))
	}
	return petri.NewPlace(name, 1, accept)
}

func (s ListPlace) Place(name string, tokenMap map[string]*petri.TokenSchema) *petri.Place {
	accepts := make([]*petri.TokenSchema, len(s))
	for i, v := range s {
		accept, ok := tokenMap[v]
		if !ok {
			panic(fmt.Sprintf("unknown token type %s", v))
		}
		accepts[i] = accept
	}
	return petri.NewPlace(name, 1, accepts...)
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
	if len(pl.AcceptedTokens) != 1 {
		panic(fmt.Errorf("place %s accepts multiple token types, and you must specify which token type to take", s))
	}
	return petri.NewArc(pl, to, pl.AcceptedTokens[0].Name, pl.AcceptedTokens[0])
}

func (s StringArc) ToPlace(net *petri.Net, from *petri.Transition) *petri.Arc {
	pl := net.Place(string(s))
	if pl == nil {
		panic(fmt.Sprintf("unknown place %s", s))
	}
	if len(pl.AcceptedTokens) != 1 {
		panic(fmt.Errorf("place %s accepts multiple token types, and you must specify which token type to put in the place", s))
	}
	return petri.NewArc(from, pl, pl.AcceptedTokens[0].Name, pl.AcceptedTokens[0])
}

type ExpressionArc map[string]string

func (e ExpressionArc) ToPlace(net *petri.Net, from *petri.Transition) *petri.Arc {
	for k, v := range e {
		pl := net.Place(k)
		if pl == nil {
			panic(fmt.Sprintf("unknown place %s", k))
		}
		for _, tok := range pl.AcceptedTokens {
			if tok.Name == v {
				return petri.NewArc(from, pl, v, tok)
			}
		}
	}
	panic(fmt.Sprintf("unknown token type %s", e))
}

func (e ExpressionArc) FromPlace(net *petri.Net, to *petri.Transition) *petri.Arc {
	for k, v := range e {
		pl := net.Place(k)
		if pl == nil {
			panic(fmt.Sprintf("unknown place %s", k))
		}
		for _, tok := range pl.AcceptedTokens {
			if tok.Name == v {
				return petri.NewArc(pl, to, v, tok)
			}
		}
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

type Petrifile struct {
	Petri       petrifile.Version
	Version     string
	Name        string
	Types       map[string]Type
	Places      map[string]interface{}
	Transitions map[string]Transition
	net         *petri.Net
	arcs        []*petri.Arc
}

func ParseInput(i interface{}) Arc {
	switch v := i.(type) {
	case string:
		return StringArc(v)
	case map[string]interface{}:
		val := make(map[string]string)
		for k, v := range v {
			val[k] = v.(string)
		}
		return ExpressionArc(val)
	}
	panic(fmt.Sprintf("unknown input type %T", i))
}

func ParseOutput(i interface{}) Arc {
	switch v := i.(type) {
	case string:
		return StringArc(v)
	case map[string]interface{}:
		val := make(map[string]string)
		for k, v := range v {
			if v == nil {
				fmt.Printf("nil value for key %s\n", k)
			}
			fmt.Printf("key %s value %v\n", k, v)
			switch value := v.(type) {
			case string:
				val[k] = value
			}
			val[k] = v.(string)
		}
		return ExpressionArc(val)
	}
	panic(fmt.Sprintf("unknown output type %T", i))
}

func PrimitiveTokenMap() map[string]*petri.TokenSchema {
	return map[string]*petri.TokenSchema{
		"string": &petri.TokenSchema{
			ID:         "",
			Type:       "",
			Properties: nil,
		},
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
		case ListPlace:
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
		if v.Event {
			t.Cold = true
		}
		ret = append(ret, t)
		if v.Inputs != nil {
			switch in := v.Inputs.(type) {
			case []interface{}:
				for _, v := range in {
					a := ParseInput(v).ToPlace(p.net, t)
					arcs = append(arcs, a)
				}
			case interface{}:
				a := ParseInput(in).ToPlace(p.net, t)
				arcs = append(arcs, a)
			}
		}

		if v.Outputs != nil {
			switch out := v.Outputs.(type) {
			case nil:
				fmt.Println("no outputs for transition %s", n)
			case []interface{}:
				for _, v := range out {
					a := ParseOutput(v).FromPlace(p.net, t)
					arcs = append(arcs, a)
				}
			case interface{}:
				a := ParseOutput(out).FromPlace(p.net, t)
				arcs = append(arcs, a)
			}
		}
	}
	p.arcs = arcs
	return ret
}

func (p *Petrifile) makeArcs() []*petri.Arc {
	return p.arcs
}

func (p *Petrifile) Net() *petri.Net {
	p.net = petri.NewNet(p.Name)
	leftover := make(map[string]Type)
	typeMap := make(map[string]map[string]petri.Properties)
	for k, v := range p.Types {
		if v.IsNested() {
			leftover[k] = v
			continue
		}
		p.net = p.net.WithTokenSchemas(v.Schema(k, typeMap))
	}
	for k, v := range leftover {
		p.net = p.net.WithTokenSchemas(v.Schema(k, typeMap))
	}
	p.net = p.net.WithPlaces(p.makePlaces()...).WithTransitions(p.makeTransitions()...).WithArcs(p.makeArcs()...)

	return p.net
}
