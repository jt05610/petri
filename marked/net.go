package marked

import (
	"errors"
	"fmt"
	"github.com/jt05610/petri"
)

type Marking []bool

var (
	ErrPlaceMarked = errors.New("place already marked")
)

type Net struct {
	*petri.Net
	marking Marking
	index   map[string]int
}

func (net *Net) Marking() Marking {
	return net.marking
}

// Enabled returns true if the transition is enabled
func (net *Net) Enabled(t *petri.Transition) bool {
	for _, arc := range net.Inputs(t) {
		if !net.marking[net.index[arc.Head.(*petri.Place).Name]] {
			return false
		}
	}
	return true
}

func (net *Net) Mark(p *petri.Place) error {
	if net.marking[net.index[p.Name]] {
		return ErrPlaceMarked
	}
	net.marking[net.index[p.Name]] = true
	return nil
}

func (net *Net) Fire(t *petri.Transition) error {
	for _, arc := range net.Inputs(t) {
		if pt, ok := arc.Head.(*petri.Place); ok {
			net.marking[net.index[pt.Name]] = false
		} else {
			return errors.New(fmt.Sprintf("%v is a transition, however only places can be inputs to transitions", arc.Head.(*petri.Transition).Name))
		}
	}
	for _, arc := range net.Outputs(t) {
		if pt, ok := arc.Tail.(*petri.Place); ok {
			net.marking[net.index[pt.Name]] = true
		}
	}
	return nil
}

func (net *Net) Available() []*petri.Transition {
	transitions := make([]*petri.Transition, 0)
	for _, t := range net.Transitions {
		if net.Enabled(t) {
			transitions = append(transitions, t)
		}
	}
	return transitions
}

func New(n *petri.Net, initial Marking) *Net {
	net := &Net{
		Net:     n,
		marking: initial,
	}
	net.index = make(map[string]int)
	for i, p := range n.Places {
		net.index[p.Name] = i
	}
	return net
}
