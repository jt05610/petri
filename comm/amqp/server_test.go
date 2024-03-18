package amqp

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"time"
)

const URL = "amqp://guest:guest@localhost:5672/"

func ExampleServer_Serve() {
	ctx, can := context.WithCancel(context.Background())
	defer can()
	dir := "../../petrifile/v1/yaml/examples"
	path := dir + "/light_switch_log.yaml"
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	bld := builder.NewBuilder(nil, dir)
	r := yaml.NewService(bld)
	bld = bld.WithService("yaml", r)
	in, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	n, err := r.Load(context.Background(), in)
	srv := NewServer(conn, n)
	m := n.NewMarking()
	for _, pl := range []string{"switch.off", "light.off"} {
		place := n.Place(pl)
		schema := place.AcceptedTokens[0]
		tok, err := schema.NewToken([]byte("hi"))
		if err != nil {
			panic(err)
		}
		err = m.PlaceTokens(place, tok)
		if err != nil {
			panic(err)
		}
	}
	srv.Initial = m
	fmt.Println(m)
	go func() {
		err = srv.Serve(ctx)
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	pubCh, err := conn.Channel()
	if err != nil {
		panic(err)

	}
	err = pubCh.PublishWithContext(ctx, "logging_light_switch", "switch.turnOn", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(""),
	})
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	err = pubCh.PublishWithContext(ctx, "logging_light_switch", "switch.turnOff", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(""),
	})
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	// Output:
	// hi
}
