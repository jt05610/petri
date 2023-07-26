package gorm

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jt05610/petri"
	"gorm.io/gorm"
)

var _ petri.Service = (*service)(nil)

type service struct {
	*gorm.DB
}

func (s *service) uuid() string {
	return uuid.New().String()
}

func Service(db *gorm.DB) petri.Service {
	return &service{db}
}

func (s *service) Nets() ([]*petri.Net, error) {
	var ret []*petri.Net
	err := s.Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *service) Net(id string) (*petri.Net, error) {
	var ret petri.Net
	err := s.First(&ret, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s *service) CreateNet(name string) (*petri.Net, error) {
	ret := &petri.Net{
		ID:   s.uuid(),
		Name: name,
	}
	err := s.Create(ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *service) DeleteNet(id string) (*petri.Net, error) {
	var ret petri.Net
	err := s.First(&ret, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = s.Delete(&ret).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s *service) AddPlace(net *petri.Net, name string, bound int) (*petri.Place, error) {
	net.Places = append(net.Places, &petri.Place{
		ID:    s.uuid(),
		Name:  name,
		Bound: bound,
	})
	err := s.Save(net).Error
	if err != nil {
		return nil, err
	}
	return net.Places[len(net.Places)-1], nil
}

func (s *service) AddTransition(net *petri.Net, name string) (*petri.Transition, error) {
	net.Transitions = append(net.Transitions, &petri.Transition{
		ID:   s.uuid(),
		Name: name,
	})
	err := s.Save(net).Error
	if err != nil {
		return nil, err
	}
	return net.Transitions[len(net.Transitions)-1], nil
}

func (s *service) AddArc(net *petri.Net, from, to petri.Node) (*petri.Arc, error) {
	net.Arcs = append(net.Arcs, &petri.Arc{
		Src:  from,
		Dest: to,
	})
	err := s.Save(net).Error
	if err != nil {
		return nil, err
	}
	return net.Arcs[len(net.Arcs)-1], nil
}

func (s *service) RemovePlace(net *petri.Net, id string) (*petri.Place, error) {
	var ret petri.Place
	err := s.First(&ret, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = s.Delete(&ret).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s *service) RemoveTransition(net *petri.Net, id string) (*petri.Transition, error) {
	var ret petri.Transition
	err := s.First(&ret, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = s.Delete(&ret).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s *service) RemoveArc(net *petri.Net, id string) (*petri.Arc, error) {
	var ret petri.Arc
	err := s.First(&ret, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = s.Delete(&ret).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s *service) UpdatePlaceName(net *petri.Net, id, name string) (*petri.Place, error) {
	pl, err := s.findPlace(net, id)
	if err != nil {
		return nil, err
	}
	pl.Name = name
	err = s.Save(pl).Error
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *service) findPlace(net *petri.Net, id string) (*petri.Place, error) {
	for _, place := range net.Places {
		if place.ID == id {
			return place, nil
		}
	}
	return nil, errors.New("place not found")
}

func (s *service) UpdatePlaceBound(net *petri.Net, id string, bound int) (*petri.Place, error) {
	pl, err := s.findPlace(net, id)
	if err != nil {
		return nil, err
	}
	pl.Bound = bound
	err = s.Save(pl).Error
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *service) findTransition(net *petri.Net, id string) (*petri.Transition, error) {
	for _, transition := range net.Transitions {
		if transition.ID == id {
			return transition, nil
		}
	}
	return nil, errors.New("transition not found")
}

func (s *service) UpdateTransitionName(net *petri.Net, id, name string) (*petri.Transition, error) {
	tr, err := s.findTransition(net, id)
	if err != nil {
		return nil, err
	}
	tr.Name = name
	err = s.Save(tr).Error
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (s *service) findArc(net *petri.Net, id string) (*petri.Arc, error) {
	for _, arc := range net.Arcs {
		if arc.ID == id {
			return arc, nil
		}
	}
	return nil, errors.New("arc not found")
}

func (s *service) UpdateArcHead(net *petri.Net, id, head string) (*petri.Arc, error) {
	arc, err := s.findArc(net, id)
	if err != nil {
		return nil, err
	}
	switch arc.Src.(type) {
	case *petri.Place:
		pl, err := s.findPlace(net, head)
		if err != nil {
			return nil, err
		}
		arc.Src = pl
	case *petri.Transition:
		tr, err := s.findTransition(net, head)
		if err != nil {
			return nil, err
		}
		arc.Src = tr
	}
	err = s.Save(arc).Error
	if err != nil {
		return nil, err
	}
	return arc, nil
}

func (s *service) UpdateArcTail(net *petri.Net, id, tail string) (*petri.Arc, error) {
	arc, err := s.findArc(net, id)
	if err != nil {
		return nil, err
	}
	switch arc.Dest.(type) {
	case *petri.Place:
		pl, err := s.findPlace(net, tail)
		if err != nil {
			return nil, err
		}
		arc.Dest = pl
	case *petri.Transition:
		tr, err := s.findTransition(net, tail)
		if err != nil {
			return nil, err
		}
		arc.Dest = tr
	}
	err = s.Save(arc).Error
	if err != nil {
		return nil, err
	}
	return arc, nil
}
