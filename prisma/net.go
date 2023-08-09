package prisma

import (
	"context"
	"errors"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/marked"
	"github.com/jt05610/petri/prisma/db"
)

type NetClient struct {
	*db.PrismaClient
	Nets                 map[string]*petri.Net
	composite            *petri.Net
	nodeIndex            map[string]petri.Node
	placeInterfaces      map[string]db.PlaceInterfaceModel
	transitionInterfaces map[string]db.TransitionInterfaceModel
	InitialMarking       map[string]int
}

func (c *NetClient) replaceInterfacePlaces(model db.PlaceInterfaceModel) error {
	// 1. add the interface place to the index and net
	newPlace := &petri.Place{
		ID:    model.ID,
		Name:  model.Name,
		Bound: model.Bound,
	}
	c.nodeIndex[model.ID] = newPlace
	c.composite.Places = append(c.composite.Places, newPlace)
	// 2. keep track of the places we are replacing, remap them, and remove them from the net
	for _, place := range model.Places() {
		c.nodeIndex[place.ID] = newPlace
		for i, p := range c.composite.Places {
			if p.ID == place.ID {
				c.composite.Places = append(c.composite.Places[:i], c.composite.Places[i+1:]...)
				break
			}
		}
		// 3. switch any arcs from this place to the interface place
		for _, arc := range c.composite.Arcs {
			if arc.Src.Identifier() == place.ID {
				arc.Src = newPlace
			}
			if arc.Dest.Identifier() == place.ID {
				arc.Dest = newPlace
			}
		}
	}
	return nil
}

func (c *NetClient) replaceInterfaceTransitions(model db.TransitionInterfaceModel) error {
	// 1. add the interface transition to the index and net
	newTransition := &petri.Transition{
		ID:   model.ID,
		Name: model.Name,
	}
	c.nodeIndex[model.ID] = newTransition
	c.composite.Transitions = append(c.composite.Transitions, newTransition)
	// 2. keep track of the transitions we are replacing and remove them from the index and net
	for _, transition := range model.Transitions() {
		c.nodeIndex[transition.ID] = newTransition
		for i, t := range c.composite.Transitions {
			if t.ID == transition.ID {
				c.composite.Transitions = append(c.composite.Transitions[:i], c.composite.Transitions[i+1:]...)
				break
			}
		}
		// 3. switch any arcs from this transition to the interface transition
		for _, arc := range c.composite.Arcs {
			if arc.Src.Identifier() == transition.ID {
				arc.Src = newTransition
			}
			if arc.Dest.Identifier() == transition.ID {
				arc.Dest = newTransition
			}
		}
	}
	return nil
}

func (c *NetClient) loadNet(ctx context.Context, id string) (*petri.Net, []string, error) {
	if c.Nets[id] != nil {
		return nil, nil, nil
	}
	net, err := c.Net.FindUnique(db.Net.ID.Equals(id)).With(db.Net.Places.Fetch()).With(db.Net.Transitions.Fetch()).With(db.Net.Arcs.Fetch()).With(db.Net.Children.Fetch()).With(db.Net.PlaceInterfaces.Fetch().With(db.PlaceInterface.Places.Fetch())).With(db.Net.TransitionInterfaces.Fetch().With(db.TransitionInterface.Transitions.Fetch())).Exec(ctx)
	if err != nil {
		return nil, nil, err
	}
	places := make([]*petri.Place, len(net.Places()))
	transitions := make([]*petri.Transition, len(net.Transitions()))
	arcs := make([]*petri.Arc, len(net.Arcs()))
	for i, place := range net.Places() {
		mark := 0
		if net.InitialMarking != nil {
			if len(net.InitialMarking) > i {
				mark = net.InitialMarking[i]
			}
		}
		places[i] = &petri.Place{
			ID:    place.ID,
			Name:  place.Name,
			Bound: place.Bound,
		}
		c.InitialMarking[place.ID] = mark
		c.nodeIndex[place.ID] = places[i]
	}
	for i, transition := range net.Transitions() {
		transitions[i] = &petri.Transition{
			ID:   transition.ID,
			Name: transition.Name,
		}
		c.nodeIndex[transition.ID] = transitions[i]
	}
	for _, pi := range net.PlaceInterfaces() {
		c.placeInterfaces[pi.ID] = pi
	}
	for _, ti := range net.TransitionInterfaces() {
		c.transitionInterfaces[ti.ID] = ti
	}
	for i, arc := range net.Arcs() {
		if arc.FromPlace {
			arcs[i] = &petri.Arc{
				ID:   arc.ID,
				Src:  c.nodeIndex[arc.PlaceID],
				Dest: c.nodeIndex[arc.TransitionID],
			}
		} else {
			arcs[i] = &petri.Arc{
				ID:   arc.ID,
				Src:  c.nodeIndex[arc.TransitionID],
				Dest: c.nodeIndex[arc.PlaceID],
			}
		}
	}
	childIDs := make([]string, len(net.Children()))
	for i, child := range net.Children() {
		childIDs[i] = child.ID
	}
	return petri.New(places, transitions, arcs, id), childIDs, nil
}

func (c *NetClient) visitChild(ctx context.Context, composite *petri.Net, id string) (*petri.Net, error) {
	if c.Nets[id] != nil {
		return nil, errors.New("already Nets")
	}
	net, childIDs, err := c.loadNet(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Nets[id] = net
	composite = petri.Add(composite, net)
	if childIDs == nil {
		return composite, nil
	}
	for _, childID := range childIDs {
		composite, err = c.visitChild(ctx, composite, childID)
		if err != nil {
			return nil, err
		}
	}
	return composite, nil
}

func (c *NetClient) Load(ctx context.Context, id string) (*marked.Net, error) {
	c.nodeIndex = make(map[string]petri.Node)
	c.Nets = make(map[string]*petri.Net)
	c.InitialMarking = make(map[string]int)
	c.composite = new(petri.Net)
	c.placeInterfaces = make(map[string]db.PlaceInterfaceModel, 0)
	c.transitionInterfaces = make(map[string]db.TransitionInterfaceModel, 0)

	var err error
	c.composite, err = c.visitChild(ctx, c.composite, id)
	if err != nil {
		return nil, err
	}
	for _, ip := range c.placeInterfaces {
		err = c.replaceInterfacePlaces(ip)
		if err != nil {
			return nil, err
		}
	}
	for _, it := range c.transitionInterfaces {
		err = c.replaceInterfaceTransitions(it)
		if err != nil {
			return nil, err
		}
	}
	return marked.NewFromMap(c.composite, c.InitialMarking), nil
}
