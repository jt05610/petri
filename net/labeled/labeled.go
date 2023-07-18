package labeled

import (
	"pnet/net"
)

type Event string

type LabeledPetriNet struct {
	*net.PetriNet
	MarkedStates map[*net.State]bool
	Labels       map[*net.Transition]Event
	InitialState *net.State
}

func LabeledNet(initial *net.State, net *net.PetriNet) *LabeledPetriNet {
	return &LabeledPetriNet{
		PetriNet:     net,
		InitialState: initial,
		MarkedStates: make(map[*net.State]bool),
		Labels:       make(map[*net.Transition]Event),
	}
}

func (lnet *LabeledPetriNet) IsMarked(state *net.State) bool {
	// Check if a state is marked
	return lnet.MarkedStates[state]
}

func (lnet *LabeledPetriNet) MarkState(state *net.State) {
	// Mark a given state
	lnet.MarkedStates[state] = true
}

func (lnet *LabeledPetriNet) Label(t *net.Transition, e Event) {
	// Label a transition with a specific event
	lnet.Labels[t] = e
}

func (lnet *LabeledPetriNet) EventFor(t *net.Transition) Event {
	// Get the event for a transition
	return lnet.Labels[t]
}
