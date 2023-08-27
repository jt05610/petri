package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strings"
	"time"
)

func snakeCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	eventMap := make(map[string]*petri.Transition)
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
	netClient := &prisma.NetClient{PrismaClient: dbClient}
	runClient := &prisma.RunClient{PrismaClient: dbClient}

	c := load(netClient, runClient, ch, exchange)
	ctx := context.Background()
	go c.Listen(ctx)
	fmt.Println()
	done := make(chan struct{})
	defer close(done)

	deviceNetIndex := make(map[string]*labeled.Net)
	seen := make(map[string]bool)
	for _, step := range c.run.Steps() {
		for _, n := range step.Action().Device().Nets() {
			if n, ok := c.nets[n.ID]; ok {
				if _, ok := seen[n.ID]; ok {
					continue
				}
				net, err := netClient.Load(ctx, n.ID)
				rawNet, err := netClient.Raw(ctx, n.ID)
				if err != nil {
					panic(err)
				}
				err = os.MkdirAll(snakeCase(step.Action().Device().Name), 0755)
				netFname := fmt.Sprintf("%s/net.json", snakeCase(step.Action().Device().Name))
				df, err := os.Create(netFname)
				if err != nil {
					panic(err)
				}
				enc := json.NewEncoder(df)
				err = enc.Encode(rawNet)
				if err != nil {
					panic(err)
				}
				err = df.Close()
				if err != nil {
					panic(err)
				}
				if err != nil {
					panic(err)
				}
				if err != nil {
					panic(err)
				}
				eventMapFname := fmt.Sprintf("%s/eventMap.json", snakeCase(snakeCase(step.Action().Device().Name)))
				for _, e := range step.Action().Device().Events() {
					for _, t := range e.Event().Transitions() {
						for _, nt := range net.Transitions {
							if t.ID == nt.ID {
								eventMap[e.Event().Name] = nt
							}
						}
					}
				}
				df, err = os.Create(eventMapFname)
				if err != nil {
					panic(err)
				}
				enc = json.NewEncoder(df)
				err = enc.Encode(eventMap)
				if err != nil {
					panic(err)
				}
				err = df.Close()
				if err != nil {
					panic(err)
				}
				deviceNetIndex[step.Action().Device().ID] = labeled.New(net)
			}
		}
	}

	expectedInitial := c.net.MarkingMap()
	// make sure devices are ready
	for _, dev := range c.Routes {
		fmt.Printf("Pinging device %s\n", dev)
		ok, err := c.Ping(ctx, dev)
		if err != nil {
			panic(err)
		}
		for k, v := range ok {
			if expectedInitial[k] != v {
				log.Fatalf("Device %s not in correct initial state", dev)
			}
		}
	}

	netCh := c.net.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-netCh:
				fmt.Println("  Client net event: ", e)
			}
		}
	}()
	for i, step := range c.run.Steps() {
		startTime := time.Now()
		err := c.Start(ctx, &step, c.events[i].Data)
		if err != nil {
			panic(err)
		}
		done := <-c.Data()
		ev := eventMap[done.Name]
		if ev == nil {
			log.Fatalf("Event %s not found", done.Name)
		}
		err = c.net.Handle(ctx, c.events[i])
		if err != nil {
			panic(err)
		}
		mainMarking := c.net.MarkingMap()
		for id, v := range done.Marking {
			mm, found := mainMarking[id]
			if !found {
				log.Fatalf("Marking for place with id %s not found", id)
			}
			if mm != v {
				placeName := ""
				for _, p := range c.net.Places {
					if p.ID == id {
						placeName = p.Name
					}
				}
				log.Fatalf("Marking for place %s (id: %s) is %d, expected %d", placeName, id, mainMarking[id], v)
			}
		}
		end := time.Now()
		fmt.Printf("Done: %s\n", done.Name)
		fmt.Printf("  From: %s\n", done.From)
		fmt.Printf("  Elapsed time: %s\n", end.Sub(startTime))
		fmt.Printf("  Data: %s\n", done.Data)
	}
}
