package graphviz

import (
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/jt05610/petri"
	"io"
)

var _ petri.Loader[*petri.Net] = (*Reader)(nil)

type Reader struct {
	*Config
	mapping     map[petri.Node]*cgraph.Node
	mappingOpp  map[string]petri.Node
	g           *cgraph.Graph
	places      []*petri.Place
	transitions []*petri.Transition
	arcs        []*petri.Arc
}

func (r *Reader) Load(reader io.Reader) (*petri.Net, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	r.g, err = cgraph.ParseBytes(bytes)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.g.Close()
	}()
	node := r.g.FirstNode()
	for node != nil {
		if node.Get("shape") == "circle" {
			p := petri.Place{
				Name: node.Get("label"),
			}
			r.places = append(r.places, &p)
			r.mapping[&p] = node
			r.mappingOpp[node.Name()] = &p
		}
		if node.Get("shape") == "box" {
			t := petri.Transition{
				Name: node.Get("label"),
			}
			r.transitions = append(r.transitions, &t)
			r.mapping[&t] = node
			r.mappingOpp[node.Name()] = &t
		}
		node = r.g.NextNode(node)
	}
	n := r.g.FirstNode()
	seen := make(map[string]bool)
	for n != nil {
		edge := r.g.FirstEdge(n)
		for edge != nil {
			if seen[edge.Name()] {
				edge = r.g.NextOut(edge)
				continue
			}
			seen[edge.Name()] = true
			other := edge.Node()
			src := r.mappingOpp[n.Name()]
			dst := r.mappingOpp[other.Name()]
			a := petri.Arc{
				Src:  src,
				Dest: dst,
			}
			r.arcs = append(r.arcs, &a)
			edge = r.g.NextOut(edge)
		}
		n = r.g.NextNode(n)
	}

	return petri.NewNet("net").WithPlaces(r.places...).WithTransitions(r.transitions...).WithArcs(r.arcs...), nil
}

func Loader() *Reader {
	return &Reader{
		mapping:     make(map[petri.Node]*cgraph.Node),
		mappingOpp:  make(map[string]petri.Node),
		places:      make([]*petri.Place, 0),
		transitions: make([]*petri.Transition, 0),
		arcs:        make([]*petri.Arc, 0),
	}
}
