package petri_test

import (
	"fmt"
	"github.com/jt05610/petri"
)

func ExampleAdd() {
	n1 := &petri.Net{
		Places: []*petri.Place{
			{Name: "a"},
			{Name: "c"},
		},
		Transitions: []*petri.Transition{
			{Name: "b"},
		},
		Arcs: []*petri.Arc{
			{Src: &petri.Place{Name: "a"}, Dest: &petri.Transition{Name: "b"}},
			{Src: &petri.Transition{Name: "b"}, Dest: &petri.Place{Name: "c"}},
		},
	}
	n2 := &petri.Net{
		Places: []*petri.Place{
			{Name: "d"},
			{Name: "c"},
		},
		Transitions: []*petri.Transition{
			{Name: "b"},
		},
		Arcs: []*petri.Arc{
			{Src: &petri.Place{Name: "d"}, Dest: &petri.Transition{Name: "b"}},
			{Src: &petri.Transition{Name: "b"}, Dest: &petri.Place{Name: "c"}},
		},
	}
	combined := petri.Add(n1, n2)
	if len(combined.Places) > 3 {
		panic("too many places")
	}
	fmt.Println("Places")
	for i, place := range combined.Places {
		if place.Name != "a" && place.Name != "c" && place.Name != "d" {
			panic("unexpected place")
		}
		fmt.Printf("%d. %s\n", i+1, place.Name)
	}

	if len(combined.Transitions) > 1 {
		panic("too many transitions")
	}
	fmt.Println("Transitions")
	fmt.Println("1. ", combined.Transitions[0].Name)

	if len(combined.Arcs) > 3 {
		panic("too many arcs")
	}
	fmt.Println("Arcs")
	for _, arc := range combined.Arcs {
		fmt.Printf("%s\n", arc)
	}
	// Output:
	// Places
	// 1. a
	// 2. c
	// 3. d
	// Transitions
	// 1.  b
	// Arcs
	// a -> b
	// b -> c
	// d -> b

}
