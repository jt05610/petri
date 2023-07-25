package petri

type NodeKind int

const (
	PlaceNode NodeKind = iota
	TransitionNode
)

type Node interface {
	Kind() NodeKind
}
