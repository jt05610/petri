package examples

import "github.com/jt05610/petri"

// Switch returns a petri net that models a switch.
func Switch() *petri.Net {
	schema := petri.Signal()
	pp := []*petri.Place{
		petri.NewPlace("off", 1, schema),
		petri.NewPlace("on", 1, schema),
	}
	tt := []*petri.Transition{
		petri.NewTransition("turnOn").WithEvent(schema),
		petri.NewTransition("turnOff").WithEvent(schema),
	}
	net := petri.NewNet("switch").WithPlaces(pp...).WithTransitions(tt...)
	aa := []*petri.Arc{
		petri.NewArc(net.Place("off"), net.Transition("turnOn"), "signal", schema),
		petri.NewArc(net.Transition("turnOn"), net.Place("on"), "signal", schema),
		petri.NewArc(net.Place("on"), net.Transition("turnOff"), "signal", schema),
		petri.NewArc(net.Transition("turnOff"), net.Place("off"), "signal", schema),
	}
	net = net.WithArcs(aa...)
	return net
}
