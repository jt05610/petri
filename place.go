package petri

var _ Object = (*Place)(nil)
var _ Node = (*Place)(nil)
var _ Input = (*PlaceInput)(nil)
var _ Filter = (*PlaceFilter)(nil)

// Place represents a place
type Place struct {
	ID string
	// Name is the name of the place
	Name string
	// Bound is the maximum number of tokens that can be in this place
	Bound int
	// AcceptedTokens are the tokens that can be accepted by this place
	AcceptedTokens []*TokenSchema
}

func (p *Place) Document() Document {
	//TODO implement me
	panic("implement me")
}

func (p *Place) From(doc Document) error {
	//TODO implement me
	panic("implement me")
}

func (p *Place) CanAccept(t *TokenSchema) bool {
	for _, token := range p.AcceptedTokens {
		if token == t {
			return true
		}
	}
	return false
}

func (p *Place) IsNode() {}

func (p *Place) Identifier() string {
	return p.Name
}

func (p *Place) String() string {
	return p.Name
}

type PlaceInput struct {
	ID    string
	Name  string
	Bound int
}

func (p *PlaceInput) Object() Object {
	//TODO implement me
	panic("implement me")
}

func (p *PlaceInput) Kind() Kind {
	return PlaceObject
}

type PlaceMask struct {
	Name  bool
	Bound bool
}

func (p *PlaceMask) IsMask() {}

type PlaceFilter struct {
	Name  string
	Bound int
	*PlaceMask
}

func (p *PlaceFilter) Filter() Document {
	//TODO implement me
	panic("implement me")
}

type PlaceUpdate struct {
	Input *PlaceInput
	Mask  *PlaceMask
}

func (p *Place) Kind() Kind { return PlaceObject }

func (p *Place) Init(i Input) error {
	in, ok := i.(*PlaceInput)
	if !ok {
		return ErrWrongInput
	}
	p.ID = in.ID
	p.Name = in.Name
	p.Bound = in.Bound
	return nil
}

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
