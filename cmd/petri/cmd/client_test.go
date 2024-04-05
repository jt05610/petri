package cmd

import (
	"testing"
)

func TestNewDependencyTree(t *testing.T) {
	input := "../../../examples/dbtl/petri/dbtl.yaml"
	net := loadNet(input)
	tree := NewDependencyGraph(net)
	if len(tree.Nodes) != 6 {
		t.Fatalf("Expected 6 node, got %d", len(tree.Nodes))
	}
	if len(tree.Nodes["dbtl"].DependsOn) != 4 {
		t.Fatalf("Expected 4 dependencies for dbtl, got %d", len(tree.Nodes["dbtl"].DependsOn))
	}
	if len(tree.Nodes["dbtl"].DependsOn["designer"].DependsOn) != 1 {
		t.Fatalf("Expected 1 dependency for designer, got %d", len(tree.Nodes["dbtl"].DependsOn["designer"].DependsOn))
	}
	nn := tree.GetNets()
	if len(nn) != 6 {
		t.Fatalf("Expected 6 nets, got %d", len(nn))
	}
	for _, n := range nn {
		t.Logf("Net: %s", n.Name)
	}

}
