package examples

import "github.com/jt05610/petri"

// Light returns a petri net that models a light.
func Light() *petri.Net {
	schema := petri.Signal()
	pp := []*petri.Place{
		petri.NewPlace("off", 1, schema),
		petri.NewPlace("shining", 1, schema),
	}
	tt := []*petri.Transition{
		petri.NewTransition("illuminate"),
		petri.NewTransition("extinguish"),
	}
	net := petri.NewNet("light").WithPlaces(pp...).WithTransitions(tt...)
	aa := []*petri.Arc{
		petri.NewArc(net.Place("off"), net.Transition("illuminate"), "signal", schema),
		petri.NewArc(net.Transition("illuminate"), net.Place("shining"), "signal", schema),
		petri.NewArc(net.Place("shining"), net.Transition("extinguish"), "signal", schema),
		petri.NewArc(net.Transition("extinguish"), net.Place("off"), "signal", schema),
	}
	net = net.WithArcs(aa...)
	return net
}
