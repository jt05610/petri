package petri_test

import (
	"github.com/jt05610/petri"
	"testing"
)

func TestToken_New(t *testing.T) {
	coin := petri.TokenSchema[interface{}]{
		Name: "Coin",
		Type: petri.Float,
	}

	penny, err := coin.NewToken(0.01)
	if err != nil {
		t.Error(err)
	}
	if penny == nil {
		t.Error("penny is nil")
	}
	shouldntWork, err := coin.NewToken("hello")

	if err == nil {
		t.Error("shouldntWork is valid")
	}
	if shouldntWork != nil {
		t.Error("shouldntWork is not nil")
	}
}
