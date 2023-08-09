package petri

func Add(nets ...*Net) *Net {
	places := make([]*Place, 0)
	seen := make(map[string]bool)
	transitions := make([]*Transition, 0)
	arcs := make([]*Arc, 0)
	for _, net := range nets {
		for _, place := range net.Places {
			if seen[place.Name] {
				continue
			}
			seen[place.Name] = true
			places = append(places, place)
		}
		for _, transition := range net.Transitions {
			if seen[transition.Name] {
				continue
			}
			seen[transition.Name] = true
			transitions = append(transitions, transition)
		}
		for _, arc := range net.Arcs {
			if seen[arc.String()] {
				continue
			}
			seen[arc.String()] = true
			arcs = append(arcs, arc)
		}
	}
	return New(places, transitions, arcs)
}
