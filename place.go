package petri

var _ Object = (*Place)(nil)
var _ Node = (*Place)(nil)
var _ Input = (*PlaceInput)(nil)
var _ Update = (*PlaceUpdate)(nil)
var _ Filter = (*PlaceFilter)(nil)

// Place represents a place
type Place struct {
	ID    string
	Name  string
	Bound int
}

func (p *Place) IsNode() {}

func (p *Place) Identifier() string {
	return p.ID
}

func (p *Place) String() string {
	return p.Name
}

type PlaceInput struct {
	Name  string
	Bound int
}

type PlaceMask struct {
	Name  bool
	Bound bool
}

type PlaceFilter struct {
	Name  string
	Bound int
	*PlaceMask
}

type PlaceUpdate struct {
	Input *PlaceInput
	Mask  *PlaceMask
}

func NewPlace(id, name string, bound int) *Place {
	return &Place{
		ID:    id,
		Name:  name,
		Bound: bound,
	}
}

func (p *Place) Kind() Kind { return PlaceObject }

func (p *Place) Init(id string, i Input) error {
	in, ok := i.(*PlaceInput)
	if !ok {
		return ErrWrongInput
	}
	p.ID = id
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
