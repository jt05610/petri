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
			{Head: pp[0], Tail: pt[0]},
			{Head: pt[0], Tail: pp[1]},
			{Head: pp[1], Tail: pt[1]},
			{Head: pt[1], Tail: pp[0]},
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
			{Head: pt[0], Tail: pt[1]},
			{Head: pt[1], Tail: pt[0]},
			{Head: pp[0], Tail: pt[0]},
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
			marking: []bool{true, false},
			seq: []*petri.Transition{
				goodNet().Transitions[0],
				goodNet().Transitions[1],
			},
			want: []bool{true, false},
			err:  nil,
		},
		{
			name:    "bad",
			net:     badNet(),
			marking: []bool{true, false},
			seq: []*petri.Transition{
				badNet().Transitions[0],
				badNet().Transitions[1],
			},
			want: []bool{true, false},
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
		{Head: pp[0], Tail: pt[0]},
		{Head: pt[0], Tail: pp[1]},
		{Head: pp[1], Tail: pt[1]},
		{Head: pt[1], Tail: pp[0]},
	}
	n := petri.New(pp, pt, aa)

	mn := marked.New(n, []bool{true, false})
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
			if !mn.Mark(p) {
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
			if !mn.Mark(p) {
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
