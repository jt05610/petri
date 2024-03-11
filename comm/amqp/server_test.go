package amqp_test

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	"os"
)

func ExampleServer_Serve() {
	dir := "../../petrifile/v1/yaml/examples"
	path := dir + "/light_switch_log.yaml"
	bld := builder.NewBuilder(nil, dir)
	r := yaml.NewService(bld)
	bld = bld.WithService("yaml", r)
	in, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	n, err := r.Load(context.Background(), in)
	m := n.NewMarking()
	for _, pl := range []string{"switch.off", "light.off"} {
		err = m.PlaceTokens(n.Place(pl), &petri.Token{
			ID:     petri.ID(),
			Schema: petri.Signal(),
			Value:  struct{}{},
		})
	}

	fmt.Println(m)

	m, err = n.Process(m, petri.Event[any]{
		Name: "switch.turnOn",
		Data: struct{}{},
	})
	fmt.Println(m)

	m, err = n.Process(m, petri.Event[any]{
		Name: "switch.turnOff",
		Data: struct{}{},
	})

	fmt.Println(m)

	// Output:
	// hi
}
