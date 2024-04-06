package petri

import (
	"errors"
	"fmt"
	"time"
)

const DequeueTimeout = 1 * time.Second

type NodeMeta struct {
	ID   string `json:"_id,omitempty"`
	Kind Kind   `json:"kind,omitempty"`
}

// Arc is a connection from a place to a transition or a transition to a place.
type Arc struct {
	ID string `json:"_id"`
	// Src is the place or transition that is the source of the arc.
	Src Node `json:"-"`
	// Dest is the place or transition that is the destination of the arc.
	Dest Node `json:"-"`
	// Expression is the expression that is evaluated when the transition connected to the arc is fired.
	Expression    string       `json:"expression,omitempty"`
	OutputSchema  *TokenSchema `json:"outputSchema,omitempty"`
	LinksNets     bool
	PlaceNet      *Net
	TransitionNet *Net
	Place         *Place
	Transition    *Transition
}

func ToValueMap(tokens map[string]Token) map[string]interface{} {
	tokenMap := make(map[string]interface{})
	for key, token := range tokens {
		tokenMap[key] = token.Value
	}
	return tokenMap
}

func AnyBytes(v any) []byte {
	switch val := v.(type) {
	case []byte:
		return val
	case string:
		return []byte(val)
	case StringValue:
		return val.Bytes()
	case SignalValue:
		return val.Bytes()
	case nil:
		return []byte{}
	default:
		panic(fmt.Errorf("cannot convert %T to []byte", v))
	}
	return nil
}

func takeFirst(tt []Token) (Token, []Token, error) {
	if len(tt) == 0 {
		return Token{}, tt, errors.New("no tokens")
	}
	if len(tt) == 1 {
		return tt[0], []Token{}, nil
	}
	return tt[0], tt[1:], nil
}

func (m Marking) pop(v string) (Token, Marking, error) {
	tt, ok := m[v]
	if !ok {
		return Token{}, m, errors.New("no tokens")
	}
	t, tt, err := takeFirst(tt)
	if err != nil {
		return Token{}, m, err
	}
	m[v] = tt
	return t, m, nil
}

func (a *Arc) TakeToken(m Marking) (Token, Marking, error) {
	if a.Src.Kind() == PlaceObject {
		pl := a.Place
		t, m, err := m.pop(pl.ID)
		if err != nil {
			return Token{}, m, err
		}
		return t, m, nil
	}
	return Token{}, m, errors.New("arc src is not a place")
}

func (m Marking) put(v string, t Token) Marking {
	tt, ok := m[v]
	if !ok {
		tt = []Token{}
	}
	tt = append(tt, t)
	m[v] = tt
	return m
}

func (a *Arc) PlaceToken(token Token, marking Marking) (Marking, error) {
	if a.Dest.Kind() == PlaceObject {
		pl := a.Place
		marking = marking.put(pl.ID, token)
		return marking, nil
	}
	return marking, errors.New("arc dest is not a place")
}

func (a *Arc) Document() Document {
	return Document{
		"_id":          a.ID,
		"src":          &NodeMeta{ID: a.Src.Identifier(), Kind: a.Src.Kind()},
		"dest":         &NodeMeta{ID: a.Dest.Identifier(), Kind: a.Dest.Kind()},
		"expression":   a.Expression,
		"outputSchema": &TokenSchema{ID: a.OutputSchema.ID},
	}
}

func NewArc(from, to Node, expression string, outputSchema *TokenSchema) *Arc {
	a := &Arc{
		ID:           ID(),
		Src:          from,
		Dest:         to,
		Expression:   expression,
		OutputSchema: outputSchema,
	}
	switch f := from.(type) {
	case *Place:
		a.Place = f
	case *Transition:
		a.Transition = f
	}
	switch t := to.(type) {
	case *Place:
		a.Place = t
	case *Transition:
		a.Transition = t
	}
	return a
}

func (a *Arc) Identifier() string {
	return a.ID
}

func (a *Arc) String() string {
	return a.Src.Identifier() + " -> " + a.Dest.Identifier()
}

func (a *Arc) Kind() Kind { return ArcObject }
