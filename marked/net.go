package marked

import (
	"errors"
	"fmt"
	"github.com/jt05610/petri"
)

type Marking []int

type Net struct {
	*petri.Net
	marking Marking
	index   map[string]int
}

func (net *Net) Copy() *Net {
	ret := &Net{
		Net:     net.Net,
		marking: make(Marking, len(net.marking)),
		index:   make(map[string]int),
	}
	for k, v := range net.index {
		ret.index[k] = v
	}
	for i, v := range net.marking {
		ret.marking[i] = v
	}
	return ret
}

func (net *Net) Marking() Marking {
	return net.marking
}

// Enabled returns true if the transition is enabled
func (net *Net) Enabled(t *petri.Transition) bool {
	for _, arc := range net.Inputs(t) {
		if pt, ok := arc.Src.(*petri.Place); ok {
			if net.marking[net.index[pt.Name]] == 0 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (net *Net) Mark(p *petri.Place) int {
	return net.marking[net.index[p.Name]]
}

var (
	TwoTransitionArc = func(s, p string) error {
		return errors.New(fmt.Sprintf("arc from %s to %s is made of two transitions", s, p))
	}
)

func (net *Net) Fire(t *petri.Transition) error {
	for _, arc := range net.Inputs(t) {
		if pt, ok := arc.Src.(*petri.Place); ok {
			net.marking[net.index[pt.Name]]--
		} else {
			head := arc.Src.(*petri.Transition)
			return TwoTransitionArc(head.Name, t.Name)
		}
	}
	for _, arc := range net.Outputs(t) {
		if pt, ok := arc.Dest.(*petri.Place); ok {
			mark := net.marking[net.index[pt.Name]]
			if mark >= pt.Bound {
				return errors.New(fmt.Sprintf("place %s is full", pt.Name))
			}
			net.marking[net.index[pt.Name]]++
		} else {
			for _, arc := range net.Inputs(t) {
				if pt, ok := arc.Src.(*petri.Place); ok {
					net.marking[net.index[pt.Name]]++
				}
			}
			tail := arc.Dest.(*petri.Transition)
			return TwoTransitionArc(t.Name, tail.Name)
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

func NewFromMap(n *petri.Net, initial map[string]int) *Net {
	marking := make(Marking, len(n.Places))
	for i, p := range n.Places {
		marking[i] = initial[p.Identifier()]
	}
	return New(n, marking)
}
