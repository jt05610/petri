package analysis

import (
	"github.com/jt05610/petri"
	"gonum.org/v1/gonum/mat"
	"strconv"
)

type Net struct {
	*petri.Net
}

type State []float64

func (net *Net) FiringVector(t int) *mat.Dense {
	v := make([]float64, len(net.Transitions))
	v[t] = 1
	return mat.NewDense(1, len(net.Transitions), v)
}

func (net *Net) arcNet(t *petri.Transition, p *petri.Place) float64 {
	ret := float64(0)
	toPlace := net.Arc(t, p)
	fromPlace := net.Arc(p, t)
	if toPlace != nil {
		ret += 1
	}
	if fromPlace != nil {
		ret -= 1
	}
	return ret

}
func (net *Net) Incidence() *mat.Dense {
	m := len(net.Places)
	n := len(net.Transitions)
	d := make([]float64, m*n)
	for i, trans := range net.Transitions {
		for j, place := range net.Places {
			d[i*m+j] = net.arcNet(trans, place)
		}
	}

	return mat.NewDense(n, m, d)
}

func (net *Net) NextState(cur map[string]bool, state *State, t *petri.Transition) (*State, bool) {
	for _, arc := range net.Inputs(t) {
		pl := arc.Head.(*petri.Place)
		if !cur[pl.Name] {
			return nil, false
		}
	}
	s := mat.NewDense(1, len(*state), *state)

	var tIndex int
	for i := range net.Transitions {
		if net.Transitions[i] == t {
			tIndex = i
			break
		}
	}

	f := net.FiringVector(tIndex)

	var result mat.Dense
	result.Mul(f, net.Incidence())

	var out mat.Dense
	out.Add(s, &result)
	ret := make(State, len(*state))
	for i := range ret {
		ret[i] = out.At(0, i)
	}
	return &ret, true
}

func (net *Net) Reachable(initial *State, target *State) bool {
	t := net.CTree(initial)
	return t.Reachable(target)
}

type TreeNode struct {
	State    *State
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

func serializeState(s *State) string {
	var ret string
	for _, i := range *s {
		v := int(i)
		var str string
		if v >= 1e6 {
			str = "ω"
		}
		str = strconv.Itoa(v)
		ret += str
	}
	return ret
}

func (net *Net) buildTree(seen map[string]bool, node *TreeNode) {
	id := serializeState(node.State)
	if _, found := seen[id]; found {
		return
	}
	seen[id] = true
	curMark := net.MappedState(node.State)
	for _, t := range net.Transitions {
		for _, pl := range net.Inputs(t) {
			if !curMark[pl.Head.String()] {
				continue
			}
		}
		next, ok := net.NextState(curMark, node.State, t)
		if !ok {
			continue
		}
		child := &TreeNode{State: next, Parent: node}
		par := child.Parent
		for par != nil {
			if (*child.State).Dominates(par.State) {
				pp := par.DominatedBy(child.State)
				for _, i := range pp {
					(*child.State)[i] = 1e6
				}
			}
			par = par.Parent
		}
		node.Children = append(node.Children, child)
	}

	for _, child := range node.Children {
		for i := range *child.State {
			if (*node.State)[i] >= 1e6 {
				(*child.State)[i] = 1e6
			}
		}
		net.buildTree(seen, child)
	}
}

func (net *Net) MappedState(s *State) map[string]bool {
	ret := make(map[string]bool)
	for i, t := range net.Places {
		ret[t.String()] = (*s)[i] > 0
	}
	return ret
}

func (net *Net) CTree(initial *State) *Tree {
	seen := make(map[string]bool)
	root := &TreeNode{State: initial}
	tree := &Tree{Root: root}
	net.buildTree(seen, tree.Root)
	return tree
}

func (t *Tree) Reachable(s *State) bool {
	ser := serializeState(s)
	for _, node := range t.Root.Children {
		if serializeState(node.State) == ser {
			return true
		}
	}
	return false
}
