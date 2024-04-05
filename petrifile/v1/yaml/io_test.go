package yaml_test

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/builder"
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
	t.Run("pump", func(t *testing.T) {
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
	})
	t.Run("logger", func(t *testing.T) {
		bld := builder.NewBuilder(nil, ".", "examples")
		r := yaml.NewService(bld)
		bld = bld.WithService("yaml", r)
		bld = bld.WithService("yml", r)
		in, err := os.Open("examples/logger.yaml")
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
		if len(n.Nets) != 0 {
			t.Error("should have 0 subnets")
		}
		m := n.NewMarking()
		for _, pl := range []string{"message"} {
			place := n.Place(pl)
			schema := place.AcceptedTokens[0]
			tok, err := schema.NewToken([]byte("hi"))
			if err != nil {
				t.Fatal(err)
			}
			err = m.PlaceTokens(n.Place(pl), tok)

			if err != nil {
				t.Fatal(err)
			}
		}

		fmt.Println(m)

		if err != nil {
			t.Error(err)
		}

		fmt.Println(m)

		//wr := graphviz.New(&graphviz.Config{
		//	Name:    "light_switch_log",
		//	Font:    graphviz.Times,
		//	RankDir: graphviz.LeftToRight,
		//})
		//out, err := os.Create("light_switch_log.svg")
		//if err != nil {
		//	t.Error(err)
		//}
		//err = wr.Flush(out, n)
		//if err != nil {
		//	t.Error(err)
		//}

	})
	t.Run("netWithSubnets", func(t *testing.T) {
		bld := builder.NewBuilder(nil, ".", "examples")
		r := yaml.NewService(bld)
		bld = bld.WithService("yaml", r)
		bld = bld.WithService("yml", r)
		in, err := os.Open("examples/light_switch_log.yaml")
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
		if n.Nets == nil {
			t.Error("nil subnets")
		}
		if len(n.Nets) != 3 {
			t.Error("wrong subnets")
		}
		m := n.NewMarking()
		for _, pl := range []string{"switch.off", "light.off"} {
			place := n.Place(pl)
			schema := place.AcceptedTokens[0]
			tok, err := schema.NewToken([]byte("hi"))
			if err != nil {
				t.Fatal(err)
			}
			err = m.PlaceTokens(n.Place(pl), tok)
			if err != nil {
				t.Fatal(err)
			}
		}

		fmt.Println(m)

		if err != nil {
			t.Error(err)
		}

		fmt.Println(m)
		if err != nil {
			t.Error(err)
		}

		fmt.Println(m)

		//wr := graphviz.New(&graphviz.Config{
		//	Name:    "light_switch_log",
		//	Font:    graphviz.Times,
		//	RankDir: graphviz.LeftToRight,
		//})
		//out, err := os.Create("light_switch_log.svg")
		//if err != nil {
		//	t.Error(err)
		//}
		//err = wr.Flush(out, n)
		//if err != nil {
		//	t.Error(err)
		//}

	})
}
