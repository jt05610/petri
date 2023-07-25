package marked_test

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/marked"
)

func ExampleNet() {
	pp := []*petri.Place{
		{Name: "closed"},
		{Name: "opened"},
	}
	pt := []*petri.Transition{
		{Name: "open"},
		{Name: "close"},
	}
	aa := []*petri.Arc{
		{Head: pp[0], Tail: pt[0]},
		{Head: pt[0], Tail: pp[1]},
		{Head: pp[1], Tail: pt[1]},
		{Head: pt[1], Tail: pp[0]},
	}
	n := &petri.Net{
		Places:      pp,
		Transitions: pt,
		Arcs:        aa,
	}
	mn := marked.New(n, []bool{true, false})
	seq := []*petri.Transition{
		pt[0],
		pt[0],
		pt[1],
		pt[1],
	}
	for _, t := range seq {
		fmt.Print("trying to fire ", t.Name)
		if !mn.Enabled(t) {
			fmt.Printf("\n  error: %s is not enabled\n", t.Name)
			continue
		}
		fmt.Print("\n  before:")
		marking := mn.Marking()
		for i, p := range mn.Places {
			if !marking[i] {
				continue
			}
			fmt.Printf(" %s", p.Name)
		}

		fmt.Print("\n")
		err := mn.Fire(t)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print("  after:")
		marking = mn.Marking()
		for i, p := range mn.Places {
			if !marking[i] {
				continue
			}
			fmt.Printf(" %s", p.Name)
		}

		fmt.Print("\n")
	}
	// Output:
	// trying to fire open
	//   before: closed
	//   after: opened
	// trying to fire open
	//   error: open is not enabled
	// trying to fire close
	//   before: opened
	//   after: closed
	// trying to fire close
	//   error: close is not enabled
}
