package labeled_test

import (
	"pnet"
	"pnet/net/labeled"
	"testing"
)

func TestLabeledNet(t *testing.T) {
	petri := petri.CaseNet()
	initial := petri.State()
	lnet := labeled.LabeledNet(&initial, petri)
	if lnet.InitialState != &initial {
		t.Errorf("initial state should be the same")
	}
	if len(lnet.MarkedStates) != 0 {
		t.Errorf("should be no marked states")
	}
	if len(lnet.Labels) != 0 {
		t.Errorf("should be no labels")
	}
}
