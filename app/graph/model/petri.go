package model

import "github.com/jt05610/petri"

type Node interface {
	IsNode()
}

type NodeType int

const (
	PlaceNode NodeType = iota
	TransitionNode
)

type Arc struct {
	ID       string `json:"id"`
	NetID    string
	HeadID   string
	HeadType NodeType
	TailID   string
	TailType NodeType
}

type Net struct {
	Places      []*Place      `json:"places"`
	Transitions []*Transition `json:"transitions"`
	Arcs        []*Arc        `json:"arcs"`
	ID          string        `json:"id"`
	Name        string        `json:"name"`
}

func (n *Net) Net() *petri.Net {
	pp := make([]*petri.Place, len(n.Places))
	nm := make(map[string]petri.Node)
	for i, p := range n.Places {
		pp[i] = p.Place
		nm[p.ID] = p.Place
	}
	tt := make([]*petri.Transition, len(n.Transitions))
	for i, t := range n.Transitions {
		tt[i] = t.Transition
		nm[t.ID] = t.Transition
	}
	aa := make([]*petri.Arc, len(n.Arcs))
	for i, a := range n.Arcs {
		aa[i] = &petri.Arc{
			Src:  nm[a.HeadID],
			Dest: nm[a.TailID],
		}
	}
	return petri.New(pp, tt, aa)
}

type NewArc struct {
	Head string `json:"head"`
	Tail string `json:"tail"`
}

type NewNet struct {
	Name        string    `json:"name"`
	Places      []*string `json:"places"`
	Transitions []*string `json:"transitions"`
	Arcs        []*NewArc `json:"arcs"`
}

type Place struct {
	*petri.Place
	NetID string
}

func (Place) IsNode() {}

type Transition struct {
	*petri.Transition
	NetID string
}

func (Transition) IsNode() {}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	NetID string
}
