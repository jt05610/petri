package marked_test

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/marked"
	"testing"
)

func TestNet_Fire(t *testing.T) {
	goodNet := func() *petri.Net {
		pp := []*petri.Place{
			{Name: "closed"},
			{Name: "opened"},
		}
		pt := []*petri.Transition{
			{Name: "open"},
			{Name: "close"},
		}
		aa := []*petri.Arc{
			{Src: pp[0], Dest: pt[0]},
			{Src: pt[0], Dest: pp[1]},
			{Src: pp[1], Dest: pt[1]},
			{Src: pt[1], Dest: pp[0]},
		}
		return petri.New(pp, pt, aa)
	}

	badNet := func() *petri.Net {
		pp := []*petri.Place{
			{Name: "closed"},
			{Name: "opened"},
		}
		pt := []*petri.Transition{
			{Name: "open"},
			{Name: "close"},
		}
		aa := []*petri.Arc{
			// Shouldn't connect two transitions
			{Src: pt[0], Dest: pt[1]},
			{Src: pt[1], Dest: pt[0]},
			{Src: pp[0], Dest: pt[0]},
		}
		return petri.New(pp, pt, aa)
	}
	testCases := []struct {
		name    string
		net     *petri.Net
		marking marked.Marking
		seq     []*petri.Transition
		want    marked.Marking
		err     error
	}{
		{
			name:    "good",
			net:     goodNet(),
			marking: []int{1, 0},
			seq: []*petri.Transition{
				goodNet().Transitions[0],
				goodNet().Transitions[1],
			},
			want: []int{1, 0},
			err:  nil,
		},
		{
			name:    "bad",
			net:     badNet(),
			marking: []int{1, 0},
			seq: []*petri.Transition{
				badNet().Transitions[0],
				badNet().Transitions[1],
			},
			want: []int{1, 0},
			err:  marked.TwoTransitionArc("open", "close"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mn := marked.New(tc.net, tc.marking)
			for _, tr := range tc.seq {
				av := mn.Available()
				for _, availTrans := range av {
					if availTrans.String() == tr.String() {
						break
					}
					t.Errorf("got %v, want %v", availTrans, tr)
				}
				err := mn.Fire(tr)
				if tc.err == nil {
					if err != nil {
						t.Errorf("got %v, want %v", err, tc.err)
					}
				}
				if tc.err != nil {
					if err == nil {
						t.Errorf("got %v, want %v", err, tc.err)
					}
				}
			}
			for i, p := range mn.Places {
				if tc.want[i] != mn.Mark(p) {
					t.Errorf("got %v, want %v", mn.Mark(mn.Places[i]), tc.want[i])
				}
			}
		})
	}
}
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
		{Src: pp[0], Dest: pt[0]},
		{Src: pt[0], Dest: pp[1]},
		{Src: pp[1], Dest: pt[1]},
		{Src: pt[1], Dest: pp[0]},
	}
	n := petri.New(pp, pt, aa)

	mn := marked.New(n, marked.Marking{1, 0})
	seq := []*petri.Transition{
		pt[0],
		pt[1],
	}
	for _, t := range seq {
		fmt.Print("trying to fire ", t.Name)
		if !mn.Enabled(t) {
			fmt.Printf("\n  error: %s is not enabled\n", t.Name)
			continue
		}
		fmt.Print("\n  before:")
		for _, p := range mn.Places {
			if mn.Mark(p) == 0 {
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
		for _, p := range mn.Places {
			if mn.Mark(p) == 0 {
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
	// trying to fire close
	//   before: opened
	//   after: closed
}
