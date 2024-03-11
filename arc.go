package petri

import (
	"errors"
	"github.com/expr-lang/expr"
)

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
	Expression   string       `json:"expression,omitempty"`
	OutputSchema *TokenSchema `json:"outputSchema,omitempty"`
}

func ToValueMap(tokens map[string]*Token) map[string]interface{} {
	tokenMap := make(map[string]interface{})
	for key, token := range tokens {
		tokenMap[key] = token.Value
	}
	return tokenMap
}

func (a *Arc) TakeToken(m Marking) (*Token, error) {
	if a.Expression != "" {
		program, err := expr.Compile(a.Expression)
		if err != nil {
			return nil, err
		}
		if a.Src.Kind() == PlaceObject {
			tokenMap := m.TokenMap(a.Src.(*Place))
			valueMap := ToValueMap(tokenMap)
			ret, err := expr.Run(program, valueMap)
			if err != nil {
				return nil, err
			}
			return a.OutputSchema.NewToken(ret)
		}
	}
	if a.Src.Kind() == PlaceObject {
		return m[a.Src.(*Place).ID].Dequeue()
	}

	return nil, errors.New("arc source is not a place")
}

func (a *Arc) PlaceToken(m Marking, tokenIndex map[string]*Token) error {
	if a.Expression != "" {
		program, err := expr.Compile(a.Expression)
		if err != nil {
			return err
		}
		if a.Dest.Kind() == PlaceObject {
			valueIndex := ToValueMap(tokenIndex)
			ret, err := expr.Run(program, valueIndex)
			if err != nil {
				return err
			}
			token, err := a.OutputSchema.NewToken(ret)
			if err != nil {
				return err
			}
			err = m.PlaceTokens(a.Dest.(*Place), token)
			if err != nil {
				return err
			}
			return nil
		}
	}
	if a.Dest.Kind() == PlaceObject {
		token, err := a.OutputSchema.NewToken(nil)
		if err != nil {
			return err
		}
		err = m.PlaceTokens(a.Dest.(*Place), token)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("arc dest is not a place")
}

func StripNodeToID(node Node) Node {
	switch node.Kind() {
	case PlaceObject:
		return &Place{
			ID: node.Identifier(),
		}
	case TransitionObject:
		return &Transition{
			ID: node.Identifier(),
		}
	default:
		return nil
	}
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

func MakeNode(k Kind, id string) Node {
	switch k {
	case PlaceObject:
		return &Place{
			ID: id,
		}
	case TransitionObject:
		return &Transition{
			ID: id,
		}
	default:
		return nil
	}
}

func NewArc(from, to Node, expression string, outputSchema *TokenSchema) *Arc {
	return &Arc{
		ID:           ID(),
		Src:          from,
		Dest:         to,
		Expression:   expression,
		OutputSchema: outputSchema,
	}
}

func (a *Arc) Identifier() string {
	return a.ID
}

func (a *Arc) String() string {
	return a.Src.Identifier() + " -> " + a.Dest.Identifier()
}

func (a *Arc) Kind() Kind { return ArcObject }
