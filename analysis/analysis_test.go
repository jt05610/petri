package analysis_test

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/analysis"
	"strings"
	"testing"
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
		{Src: pp[0], Dest: tt[0]},
		{Src: tt[0], Dest: pp[1]},
		{Src: pp[1], Dest: tt[1]},
		{Src: tt[1], Dest: pp[2]},
		{Src: pp[2], Dest: tt[0]},
		{Src: tt[1], Dest: pp[3]},
		{Src: pp[3], Dest: tt[2]},
		{Src: tt[2], Dest: pp[0]},
	}
	net := petri.NewNet(pp, tt, aa)
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

func TestNet_Reachable(t *testing.T) {
	n := net()
	for _, tc := range []struct {
		name string
		from *analysis.State
		to   *analysis.State
		want bool
	}{
		{
			name: "Negative solution value",
			from: &analysis.State{1, 0, 0, 0},
			to:   &analysis.State{0, 0, 0, 1},
			want: false,
		},
		{
			name: "No solution",
			from: &analysis.State{1, 0, 0, 0},
			to:   &analysis.State{0, 1, 0, 0},
			want: false,
		},
		{
			name: "Happy",
			from: &analysis.State{1, 0, 1, 0},
			to:   &analysis.State{0, 1, 0, 0},
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := n.Reachable(tc.from, tc.to)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNet_CTree(t *testing.T) {
	nPlaces := 4
	nTransitions := 3
	pp := make([]*petri.Place, nPlaces)
	for i := 0; i < nPlaces; i++ {
		pp[i] = &petri.Place{Name: fmt.Sprintf("p%d", i+1)}
	}
	tt := make([]*petri.Transition, nTransitions)
	for i := 0; i < nTransitions; i++ {
		tt[i] = &petri.Transition{
			Name: fmt.Sprintf("t%d", i+1)}
	}
	aa := []*petri.Arc{
		{Src: pp[0], Dest: tt[0]},
		{Src: tt[0], Dest: pp[1]},
		{Src: pp[1], Dest: tt[1]},
		{Src: tt[1], Dest: pp[0]},
		{Src: pp[1], Dest: tt[2]},
		{Src: tt[2], Dest: pp[2]},
		{Src: pp[2], Dest: tt[2]},
		{Src: tt[0], Dest: pp[2]},
		{Src: tt[2], Dest: pp[3]},
	}
	initial := &analysis.State{1, 0, 0, 0}
	n := petri.NewNet(pp, tt, aa)
	aNet := &analysis.Net{Net: n}
	ct := aNet.CTree(initial)
	if ct == nil {
		t.Errorf("ctree is nil")
	}
	if ct.Root == nil {
		t.Errorf("ctree root is nil")
	}
	if ct.Root.State != initial {
		t.Errorf("ctree root state is not initial state")
	}
}
