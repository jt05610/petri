package examples

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
)

func Net() *petri.Net {
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
		{Src: tt[0], Dest: pp[2]},
		{Src: pp[1], Dest: tt[1]},
		{Src: tt[1], Dest: pp[0]},
		{Src: pp[1], Dest: tt[2]},
		{Src: pp[2], Dest: tt[2]},
		{Src: tt[2], Dest: pp[2]},
		{Src: tt[2], Dest: pp[3]},
	}
	return petri.NewNet(pp, tt, aa)
}

func LabeledNet(m marked.Marking) *labeled.Net {
	mn := marked.New(Net(), m)
	return labeled.New(mn)
}
