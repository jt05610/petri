package place_test

import (
	"pnet/place"
	"testing"
)

type testToken struct{}

func (t testToken) IsToken() {}

func TestUnboundedPlace(t *testing.T) {
	pl := place.NewBuilder[*testToken]("p1").Build()
	if pl.Name() != "p1" {
		t.Errorf("name should be p1")
	}
	if pl.ID() == "" {
		t.Errorf("id should not be empty")
	}
	if len(pl.Inputs(nil)) != 0 {
		t.Errorf("inputs should be empty")
	}
	if len(pl.Outputs(nil)) != 0 {
		t.Errorf("outputs should be empty")
	}
	for i := 0; i < 1000; i++ {
		err := pl.Add(&testToken{})
		if err != nil {
			t.Errorf("should not error when adding tokens")
		}
	}
	out, err := pl.Pop(1000)
	if err != nil {
		t.Errorf("should not error when popping tokens")
	}
	if len(out) != 1000 {
		t.Errorf("should have popped 1000 tokens")
	}
	_, err = pl.Pop(1)
	if err == nil {
		t.Errorf("should error when popping more tokens than are available")
	}
	if err != place.ErrNotEnoughTokens {
		t.Errorf("should be ErrNotEnoughTokens")
	}
}
