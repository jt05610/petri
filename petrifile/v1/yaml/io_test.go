package yaml_test

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/examples"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	"os"
	"testing"
)

func TestService_Save(t *testing.T) {
	n := examples.NewPump().Net
	w := &yaml.Service{}
	out, err := os.Create("test.petri")
	if err != nil {
		t.Error(err)
	}
	err = w.Save(nil, out, n)
	if err != nil {
		t.Error(err)
	}
}

func TestService_Read(t *testing.T) {
	r := &yaml.Service{}
	in, err := os.Open("examples/pump.yaml")
	if err != nil {
		t.Error(err)
	}
	n, err := r.Load(context.Background(), in)
	if err != nil {
		t.Error(err)
	}
	if n == nil {
		t.Error("nil net")
	}
	if n.Name != "pump" {
		t.Error("wrong name")
	}
	if len(n.TokenSchemas) != 5 {
		t.Error("wrong token schemas")
		fmt.Println(n.TokenSchemas)
	}
	if len(n.Places) != 5 {
		t.Error("wrong places")
	}
	if len(n.Transitions) != 4 {
		t.Error("wrong transitions")
	}
	if len(n.Arcs) == 0 {
		t.Error("wrong arcs")
	}
}
