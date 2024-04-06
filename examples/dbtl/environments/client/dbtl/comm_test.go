package main

import (
	"client/model/proto/v1/model"
	"github.com/jt05610/petri"
	"math/rand"
	"testing"
)

func randomValue() *float32 {
	v := rand.Float32()
	return &v
}

func TestTokensToProto(t *testing.T) {
	fields := "abc"
	unit := "m"
	values := []*model.Quantity{
		{
			Value: randomValue(),
			Unit:  &unit,
		},
	}
	name := "sam"
	expected := &model.Sample{
		Fields: &fields,
		Name:   &name,
		Values: values,
	}
	tokens := map[string]petri.Token{
		"fields": {
			Value: []byte(fields),
		},
	}
}
