package analysis_test

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/analysis"
	"strings"
)

func net() *analysis.Net {
	pp := make([]*petri.Place, 4)
	for i := 0; i < 4; i++ {
		pp[i] = &petri.Place{Name: fmt.Sprintf("p%d", i+1)}
	}
	tt := make([]*petri.Transition, 3)
	for i := 0; i < 3; i++ {
		tt[i] = &petri.Transition{
			Name: fmt.Sprintf("t%d", i+1)}
	}
	aa := []*petri.Arc{
		{Head: pp[0], Tail: tt[0]},
		{Head: tt[0], Tail: pp[1]},
		{Head: pp[1], Tail: tt[1]},
		{Head: tt[1], Tail: pp[2]},
		{Head: pp[2], Tail: tt[0]},
		{Head: tt[1], Tail: pp[3]},
		{Head: pp[3], Tail: tt[2]},
		{Head: tt[2], Tail: pp[0]},
	}
	net := petri.New(pp, tt, aa)
	return &analysis.Net{Net: net}
}
func ExampleNet_Incidence() {
	aNet := net()
	inc := aNet.Incidence()
	fmt.Printf("┌%s┐\n", strings.Repeat(" ", 3*len(aNet.Places)-1))
	for i := range aNet.Transitions {
		fmt.Print("│")
		s := " "
		for j := range aNet.Places {
			if j == len(aNet.Places)-1 {
				s = ""
			}
			fmt.Printf("%2d%s", int(inc.At(i, j)), s)
		}
		fmt.Print("│\n")
	}
	fmt.Printf("└%s┘", strings.Repeat(" ", 3*len(aNet.Places)-1))
	// Output:
	// ┌           ┐
	// │-1  1 -1  0│
	// │ 0 -1  1  1│
	// │ 1  0  0 -1│
	// └           ┘
}
