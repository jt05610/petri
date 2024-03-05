package marked

import (
	"errors"
	"fmt"
	"github.com/jt05610/petri"
)

type Marking []int

type Net struct {
	*petri.Net
	marking      Marking
	index        map[string]int
	joinedPlaces map[string]string
}

func (n *Net) Copy() *Net {
	ret := &Net{
		Net:     n.Net,
		marking: make(Marking, len(n.marking)),
		index:   make(map[string]int),
	}
	for k, v := range n.index {
		ret.index[k] = v
	}
	for i, v := range n.marking {
		ret.marking[i] = v
	}
	return ret
}

func (n *Net) MarkingMap() map[string]int {
	ret := make(map[string]int)
	for k, v := range n.index {
		ret[k] = n.marking[v]
	}
	for ind, joined := range n.joinedPlaces {
		ret[ind] = n.marking[n.index[joined]]
	}
	return ret
}

func (n *Net) Marking() Marking {
	return n.marking
}

// Enabled returns true if the transition is enabled
func (n *Net) Enabled(t *petri.Transition) bool {
	for _, arc := range n.Inputs(t) {
		if pt, ok := arc.Src.(*petri.Place); ok {
			if n.marking[n.index[pt.ID]] == 0 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (n *Net) Mark(p *petri.Place) int {
	return n.marking[n.index[p.Name]]
}

var (
	TwoTransitionArc = func(s, p string) error {
		return errors.New(fmt.Sprintf("arc from %s to %s is made of two transitions", s, p))
	}
)

func (n *Net) Fire(t *petri.Transition) error {
	for _, arc := range n.Inputs(t) {
		if pt, ok := arc.Src.(*petri.Place); ok {
			n.marking[n.index[pt.ID]]--
		} else {
			head := arc.Src.(*petri.Transition)
			return TwoTransitionArc(head.Name, t.Name)
		}
	}
	for _, arc := range n.Outputs(t) {
		if pt, ok := arc.Dest.(*petri.Place); ok {
			mark := n.marking[n.index[pt.ID]]
			if mark >= pt.Bound {
				// ignore
				continue
			}
			n.marking[n.index[pt.ID]]++
		} else {
			for _, arc := range n.Inputs(t) {
				if pt, ok := arc.Src.(*petri.Place); ok {
					n.marking[n.index[pt.ID]]++
				}
			}
			tail := arc.Dest.(*petri.Transition)
			return TwoTransitionArc(t.Name, tail.Name)
		}
	}
	return nil
}

func (n *Net) Available() []*petri.Transition {
	transitions := make([]*petri.Transition, 0)
	for _, t := range n.Transitions {
		if n.Enabled(t) {
			transitions = append(transitions, t)
		}
	}
	return transitions
}

func New(n *petri.Net, initial Marking, joinedIDs ...map[string]string) *Net {
	net := &Net{
		Net:     n,
		marking: initial,
	}
	net.joinedPlaces = make(map[string]string)
	for _, joined := range joinedIDs {
		for k, v := range joined {
			net.joinedPlaces[k] = v
		}
	}
	return net
}

func NewFromMap(n *petri.Net, initial map[string]int, joinedIDs ...map[string]string) *Net {
	marking := make(Marking, len(n.Places))
	return New(n, marking, joinedIDs...)
}
