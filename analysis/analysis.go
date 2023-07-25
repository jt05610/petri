package analysis

import (
	"github.com/jt05610/petri"
	"gonum.org/v1/gonum/mat"
)

type Net struct {
	*petri.Net
}

type State []float64

func (s *State) Clone() *State {
	clone := make(State, len(*s))
	for place, tokens := range *s {
		clone[place] = tokens
	}
	return &clone
}

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

func (net *Net) NextState(state *State, t *petri.Transition) (*State, bool) {
	for i := range net.Inputs(t) {
		if (*state)[i] == 0 {
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
	in := mat.NewDense(1, len(*initial), *initial)
	res := mat.NewDense(1, len(*target), *target)
	res.Sub(res, in)
	inc := net.Incidence()
	var sol mat.Dense
	err := sol.Solve(res.T(), inc.T())
	if err != nil {
		return false
	}
	for i, _ := range net.Transitions {
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

func (net *Net) buildTree(node *TreeNode) {
	found := false
	for _, t := range net.Transitions {
		for i := range net.Inputs(t) {
			if (*node.State)[i] > 0 {
				found = true
				next, ok := net.NextState(node.State, t)
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
					net.buildTree(child)
				}
			}
		}
		if !found {
			node.mark = Terminal
		}
	}
}

func (net *Net) CTree(initial *State) *Tree {
	root := &TreeNode{State: initial}
	tree := &Tree{Root: root}
	net.buildTree(tree.Root)
	return tree
}
