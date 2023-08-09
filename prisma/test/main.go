package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/amqp/client"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"time"
)

func event(model *db.EventModel) *labeled.Event {
	return &labeled.Event{
		Name: model.Name,
		Data: model.Data,
	}
}

func makeHandler(eN string) labeled.Handler {
	return func(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
		fmt.Printf("Handling event %s with handler %s\n", event.Name, eN)
		return event, nil
	}
}

type Controller struct {
	*client.Controller
	net *labeled.Net
	run *db.RunModel
}

func NewController(ch *amqp.Channel, exchange string, net *labeled.Net, model *db.RunModel, routes map[string]string) *Controller {
	return &Controller{
		Controller: client.NewController(ch, exchange, routes),
		net:        net,
		run:        model,
	}
}

func load(dbClient *db.PrismaClient, ch *amqp.Channel, exchange string, routes ...map[string]string) *Controller {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := &prisma.RunClient{PrismaClient: dbClient}

	res, err := srv.List(ctx)
	if err != nil {
		panic(err)
	}

	fst := res[0]
	run, err := srv.Load(ctx, fst.ID)
	if err != nil {
		panic(err)
	}

	netClient := &prisma.NetClient{PrismaClient: dbClient}

	net, err := netClient.Load(ctx, run.NetID)

	ln := labeled.New(net)

	events := make([]*labeled.Event, len(run.Steps()))

	var rm map[string]string
	if len(routes) == 0 {
		rm = make(map[string]string)
	} else {
		rm = routes[0]
	}

	for i, step := range run.Steps() {
		events[i] = event(step.Action().Event())
		step.Action().Event().Transitions()
		fmt.Println(events[i])
		if len(routes) == 0 {
			rm[step.Action().Device().ID] = step.Action().Device().Instances()[0].ID
		}
		for _, t := range step.Action().Event().Transitions() {
			for _, nt := range net.Transitions {
				if nt.ID == t.ID {
					err := ln.AddHandler(events[i].Name, nt, makeHandler(events[i].Name))
					if err != nil {
						panic(err)
					}
				}
			}
		}
		if err != nil {
			panic(err)
		}
	}

	ok := labeled.ValidSequence(ln, events)
	if !ok {
		panic("Invalid sequence")
	}
	fmt.Println("\n\nValid sequence")
	for i, s := range run.Steps() {
		fmt.Printf("\nStep %d\n", i+1)
		fmt.Printf("  Action: %s\n", s.Action().Event().Name)
		fmt.Printf("  Device: %s\n", s.Action().Device().Name)
		fmt.Printf("  Addr: %s\n", s.Action().Device().Instances()[0].Addr)
	}
	return NewController(ch, exchange, ln, run, rm)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	exchange := os.Getenv("AMQP_EXCHANGE")
	dbClient := db.NewClient()
	if err := dbClient.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := dbClient.Disconnect(); err != nil {
			panic(err)
		}
	}()
	uri := os.Getenv("RABBITMQ_URI")
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	c := load(dbClient, ch, exchange)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Listen(ctx)
	fmt.Println()
	for _, step := range c.run.Steps() {
		startTime := time.Now()
		err := c.Start(ctx, &step)
		if err != nil {
			panic(err)
		}
		done := <-c.Data()
		end := time.Now()
		fmt.Printf("Done: %s\n", done.Name)
		fmt.Printf("  From: %s\n", done.From)
		fmt.Printf("  Elapsed time: %s\n", end.Sub(startTime))
		fmt.Printf("  Data: %s\n", done.Data)
	}
}
