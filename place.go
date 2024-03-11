package petri

import "errors"

var _ Object = (*Place)(nil)
var _ Node = (*Place)(nil)
var _ Input = (*PlaceInput)(nil)
var _ Filter = (*PlaceFilter)(nil)

// Place represents a place.
type Place struct {
	ID string `json:"_id"`
	// Name is the name of the place
	Name string `json:"name,omitempty"`
	// Bound is the maximum number of tokens that can be in this place
	Bound int `json:"bound,omitempty"`
	// AcceptedTokens are the tokens that can be accepted by this place
	AcceptedTokens []*TokenSchema `json:"acceptedTokens,omitempty"`
	TokenQueue
}

func (p *Place) Index() string {
	return p.Name
}

func (p *Place) PostInit() error {
	return nil
}

// NewPlace creates a new place.
func NewPlace(name string, bound int, acceptedTokens ...*TokenSchema) *Place {
	return &Place{
		ID:             ID(),
		Name:           name,
		Bound:          bound,
		AcceptedTokens: acceptedTokens,
		TokenQueue:     NewLocalQueue(bound),
	}
}

func acceptedTokenIDs(tokens []*TokenSchema) []*TokenSchema {
	ids := make([]*TokenSchema, len(tokens))
	for i, token := range tokens {
		ids[i] = &TokenSchema{
			ID: token.ID,
		}
	}
	return ids
}

func (p *Place) Document() Document {
	return Document{
		"_id":            p.ID,
		"name":           p.Name,
		"bound":          p.Bound,
		"acceptedTokens": acceptedTokenIDs(p.AcceptedTokens),
	}
}

func (p *Place) CanAccept(t *TokenSchema) bool {
	for _, token := range p.AcceptedTokens {
		if token.Name == t.Name {
			return true
		}
	}
	return false
}

func (p *Place) IsNode() {}

func (p *Place) Identifier() string {
	return p.ID
}

func (p *Place) String() string {
	return p.Name
}

type PlaceInput struct {
	Name           string
	Bound          int
	AcceptedTokens []*TokenSchema
}

func (p *PlaceInput) Object() Object {
	return NewPlace(p.Name, p.Bound, p.AcceptedTokens...)
}

func (p *PlaceInput) Kind() Kind {
	return PlaceObject
}

type PlaceMask struct {
	Name           bool
	Bound          bool
	AcceptedTokens bool
}

var ErrNotFound = errors.New("not found")

func (p *PlaceMask) IsMask() {}

type NodeFilter interface {
	Filter
	IsNodeFilter()
}

type PlaceFilter struct {
	ID    *StringSelector `json:"_id,omitempty"`
	Name  *StringSelector `json:"name,omitempty"`
	Bound *IntSelector    `json:"bound,omitempty"`
}

func (p *PlaceFilter) IsNodeFilter() {}

type PlaceUpdate struct {
	Input *PlaceInput
	Mask  *PlaceMask
}

func (p *Place) Kind() Kind { return PlaceObject }

func (p *Place) Update(u Update) error {
	update, ok := u.(*PlaceUpdate)
	if !ok {
		return ErrWrongUpdate
	}
	if update.Mask.Name {
		p.Name = update.Input.Name
	}
	if update.Mask.Bound {
		p.Bound = update.Input.Bound
	}
	return nil
}

func (p *PlaceInput) IsInput()   {}
func (p *PlaceUpdate) IsUpdate() {}
func (p *PlaceFilter) IsFilter() {}
