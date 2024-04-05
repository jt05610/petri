package petri

import (
	"context"
	"errors"
	"fmt"
	"github.com/expr-lang/expr"
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

func (a *Arc) TakeToken() (Token, error) {
	if a.Expression != "" {
		program, err := expr.Compile(a.Expression)
		if err != nil {
			return Token{}, err
		}
		if a.Src.Kind() == PlaceObject {
			ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
			defer can()
			pl := a.Place
			tok, err := pl.TokenQueue.Dequeue(ctx)
			if err != nil {
				return Token{}, err
			}
			tokMap := map[string]Token{
				tok.Schema.Name: tok,
			}
			valueMap := ToValueMap(tokMap)
			ret, err := expr.Run(program, valueMap)
			if err != nil {
				return Token{}, err
			}
			return a.OutputSchema.NewToken(AnyBytes(ret))
		}
	}
	if a.Src.Kind() == PlaceObject {
		ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
		defer can()
		pl := a.Place
		tok, err := pl.TokenQueue.Dequeue(ctx)
		if err != nil {
			return Token{}, err
		}
		return tok, nil
	}

	return Token{}, errors.New("arc source is not a place")
}

func (a *Arc) PlaceToken(tokenIndex map[string]Token) error {
	if a.Expression != "" {
		program, err := expr.Compile(a.Expression)
		if err != nil {
			return err
		}
		if a.Dest.Kind() == PlaceObject {
			ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
			defer can()
			valueIndex := ToValueMap(tokenIndex)
			ret, err := expr.Run(program, valueIndex)
			if err != nil {
				return err
			}
			token, err := a.OutputSchema.NewToken(AnyBytes(ret))
			if err != nil {
				return err
			}
			err = a.Place.TokenQueue.Enqueue(ctx, token)
			if err != nil {
				return err
			}
			return nil
		}
	}
	if a.Dest.Kind() == PlaceObject {
		ctx, can := context.WithTimeout(context.Background(), DequeueTimeout)
		defer can()
		token, ok := tokenIndex[a.OutputSchema.Name]
		if !ok {
			return errors.New("token not found")
		}
		err := a.Place.TokenQueue.Enqueue(ctx, token)
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
