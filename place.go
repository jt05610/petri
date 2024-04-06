package petri

import "errors"

var _ Node = (*Place)(nil)

// Place represents a place.
type Place struct {
	ID string `json:"_id"`
	// Name is the name of the place
	Name string `json:"name,omitempty"`
	// Bound is the maximum number of tokens that can be in this place
	Bound int `json:"bound,omitempty"`
	// AcceptedTokens are the tokens that can be accepted by this place
	Schema *TokenSchema
	TokenQueue
	// IsEvent is true if this place represents an event
	IsEvent bool
}

func (p *Place) Index() string {
	return p.Name
}

func (p *Place) PostInit() error {
	return nil
}

// NewPlace creates a new place.
func NewPlace(name string, bound int, acceptedTokens *TokenSchema) *Place {
	return &Place{
		ID:         ID(),
		Name:       name,
		Bound:      bound,
		Schema:     acceptedTokens,
		TokenQueue: NewLocalQueue(bound),
	}
}

func (p *Place) CanAccept(t *TokenSchema) bool {
	if p.Schema == nil {
		return false
	}
	if p.Schema.Name == t.Name {
		return true
	}
	if p.Schema.ID == t.ID {
		return true
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
