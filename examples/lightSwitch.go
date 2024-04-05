package examples

import "github.com/jt05610/petri"

func LightSwitch() *petri.Net {
	net := petri.NewNet("lightSwitch").WithNets(Switch(), Light())
	signal := petri.Signal()
	aa := []*petri.Arc{
		petri.NewArc(net.Place("switch.on"), net.Transition("light.illuminate"), "signal", signal),
		petri.NewArc(net.Transition("light.illuminate"), net.Place("switch.on"), "signal", signal),
		petri.NewArc(net.Place("switch.off"), net.Transition("light.extinguish"), "signal", signal),
		petri.NewArc(net.Transition("light.extinguish"), net.Place("switch.off"), "signal", signal),
	}
	net = net.WithArcs(aa...)
	return net
}
