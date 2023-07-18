package net

import (
	"context"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"reflect"
)

// Arc that connects places and transitions, with a weight
type Arc struct {
	Head Node
	Tail Node
}

// Transition struct
type Transition struct {
	Name   string
	Handle func(ctx context.Context, args ...Token) []Token
}

func (t *Transition) IsNode() {}

func (p *Place) Accepts(t Token) bool {
	fmt.Print(reflect.TypeOf(t))
	if reflect.TypeOf(p.accepts) == reflect.TypeOf(t) {
		return true
	}
	return false
}

func (p *Place) IsNode() {}

func (p *Place) Add(t ...Token) {
	p.tokens = append(p.tokens, t...)
}

func (p *Place) Pop(n int) []Token {
	if n > len(p.tokens) {
		n = len(p.tokens)
	}
	tokens := p.tokens[:n]
	p.tokens = p.tokens[n:]
	return tokens
}

type Places []*Place

type Transitions []*Transition

type Marking []int

// PetriNet struct
type PetriNet struct {
	Places      Places
	Transitions Transitions
	Arcs        []*Arc
	Weights     map[*Arc]int
	ArcIndexes  map[Node]map[Node]*Arc
}

func (p *PetriNet) LoopingWeightOf(head, tail Node) int {
	for _, arc := range p.Arcs {
		if arc.Head == head && arc.Tail == tail {
			return p.Weights[arc]
		}
	}
	return 0
}

func (p *PetriNet) WeightOf(head, tail Node) int {
	if arc, ok := p.ArcIndexes[head][tail]; ok {
		return p.Weights[arc]
	}
	return 0
}

func (p *PetriNet) Arc(tail, head Node) *Arc {
	for _, arc := range p.Arcs {
		if arc.Tail == tail && arc.Head == head {
			return arc
		}
	}
	return nil
}

func (p *PetriNet) Inputs(n Node) []*Arc {
	var inputs []*Arc
	for _, arc := range p.Arcs {
		if arc.Tail == n {
			inputs = append(inputs, arc)
		}
	}
	return inputs
}

func (p *PetriNet) Outputs(n Node) []*Arc {
	var outputs []*Arc
	for _, arc := range p.Arcs {
		if arc.Head == n {
			outputs = append(outputs, arc)
		}
	}
	return outputs
}

// Enabled returns true if the transition is enabled
func (p *PetriNet) Enabled(t *Transition) bool {
	for _, arc := range p.Inputs(t) {
		if len(arc.Head.(*Place).tokens) < p.Weights[arc] {
			return false
		}
	}
	return true
}

type State []float64

func (x *State) Clone() *State {
	clone := make(State, len(*x))
	for place, tokens := range *x {
		clone[place] = tokens
	}
	return &clone
}
func (p *PetriNet) State() State {
	m := make(State, len(p.Places))
	for i, place := range p.Places {
		m[i] = float64(len(place.tokens))
	}
	return m
}

func NotEnabled(t string) error {
	return fmt.Errorf("transition %s is not enabled", t)
}
func (p *PetriNet) Fire(ctx context.Context, e int) error {
	transition := p.Transitions[e]

	inputs := p.Inputs(transition)
	in := make([]Token, len(inputs))
	for _, arc := range inputs {
		pl := arc.Head.(*Place)
		in = append(in, pl.Pop(p.Weights[arc])...)
	}
	out := transition.Handle(ctx, in...)
	for i, arc := range p.Outputs(transition) {
		pl := arc.Tail.(*Place)
		if pl.Accepts(out[i]) {
			pl.Add(out[i])
		}
	}
	return nil
}

func (p *PetriNet) Available() Transitions {
	transitions := make(Transitions, 0)
	for _, t := range p.Transitions {
		if p.Enabled(t) {
			transitions = append(transitions, t)
		}
	}
	return transitions
}

func (p *PetriNet) FiringVector(t int) *mat.Dense {
	v := make([]float64, len(p.Transitions))
	v[t] = 1
	return mat.NewDense(1, len(p.Transitions), v)
}

func (p *PetriNet) Incidence() *mat.Dense {
	m := len(p.Places)
	n := len(p.Transitions)
	d := make([]float64, m*n)
	for i, trans := range p.Transitions {
		for j, place := range p.Places {
			d[i*m+j] = float64(p.WeightOf(trans, place) - p.WeightOf(place, trans))
		}
	}

	return mat.NewDense(n, m, d)
}

func (p *PetriNet) NextState(state *State, t *Transition) (*State, bool) {
	for i := range p.Inputs(t) {
		if (*state)[i] < float64(p.Weights[p.Inputs(t)[i]]) {
			return nil, false
		}
	}
	s := mat.NewDense(1, len(*state), *state)

	var tIndex int
	for i := range p.Transitions {
		if p.Transitions[i] == t {
			tIndex = i
			break
		}
	}

	f := p.FiringVector(tIndex)

	var result mat.Dense
	result.Mul(f, p.Incidence())

	var out mat.Dense
	out.Add(s, &result)
	ret := make(State, len(*state))
	for i := range ret {
		ret[i] = out.At(0, i)
	}
	return &ret, true
}

func (p *PetriNet) Reachable(initial *State, target *State) bool {
	in := mat.NewDense(1, len(*initial), *initial)
	res := mat.NewDense(1, len(*target), *target)
	res.Sub(res, in)
	inc := p.Incidence()
	var sol mat.Dense
	err := sol.Solve(res.T(), inc.T())
	if err != nil {
		return false
	}
	for i, _ := range p.Transitions {
		if sol.At(0, i) < 0 {
			return false
		}
	}
	return true
}

type TreeNodeMark int

const (
	Unmarked TreeNodeMark = iota
	Terminal
	Duplicate
)

type TreeNode struct {
	State    *State
	mark     TreeNodeMark
	Parent   *TreeNode
	Children []*TreeNode
}

func (s *State) Dominates(b *State) bool {
	oneGt := false
	for i := range *s {
		if (*s)[i] < (*b)[i] {
			return false
		}
		if (*s)[i] > (*b)[i] {
			oneGt = true
		}
	}
	return oneGt
}

func (t *TreeNode) DominatedBy(s *State) []int {
	var dominating []int
	for i := range *s {
		if (*s)[i] > (*t.State)[i] {
			dominating = append(dominating, i)
		}
	}
	if len(dominating) != 0 {
		return dominating
	}
	if t.Parent != nil {
		return t.Parent.DominatedBy(s)
	}
	return nil
}

type Tree struct {
	Root *TreeNode
}

func (p *PetriNet) buildTree(node *TreeNode) {
	found := false
	for _, t := range p.Transitions {
		for i := range p.Inputs(t) {
			if (*node.State)[i] >= float64(p.Weights[p.Inputs(t)[i]]) {
				found = true
				next, ok := p.NextState(node.State, t)
				if !ok {
					continue
				}
				same := true
				for i := range *node.State {
					if (*node.State)[i] == (*next)[i] {
						continue
					}
					if (*node.State)[i] > 1e6 {
						(*next)[i] = 1e6
					}
				}
				if same {
					node.mark = Duplicate
					return
				}
				if node.DominatedBy(next) != nil {
					for _, i := range node.DominatedBy(next) {
						(*next)[i] = 1e6
					}
					child := &TreeNode{State: next}
					node.Children = append(node.Children, child)
					p.buildTree(child)
				}
			}
		}
		if !found {
			node.mark = Terminal
		}
	}
}

func (p *PetriNet) CTree(initial *State) *Tree {
	root := &TreeNode{State: initial}
	tree := &Tree{Root: root}
	p.buildTree(tree.Root)
	return tree
}
