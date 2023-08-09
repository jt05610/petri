package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/amqp/client"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func fLower(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func event(model *db.EventModel) *labeled.Event {
	fields := make([]*labeled.Field, len(model.Fields()))
	for i, f := range model.Fields() {
		fields[i] = &labeled.Field{
			Name: fLower(f.Name),
			Type: labeled.FieldType(f.Type),
		}
	}
	return &labeled.Event{
		Name:   model.Name,
		Data:   nil,
		Fields: fields,
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
	nets   map[string]*labeled.Net
	net    *labeled.Net
	run    *db.RunModel
	events []*labeled.Event
}

func echoHandlers(n *labeled.Net) control.Handlers {
	handlers := make(control.Handlers)
	for _, e := range n.Events {
		handlers[snakeCase(e.Name)] = makeHandler(e.Name)
	}
	return handlers
}

func NewController(events []*labeled.Event, ch *amqp.Channel, exchange string, nets map[string]*labeled.Net, net *labeled.Net, model *db.RunModel, routes map[string]string) *Controller {
	return &Controller{
		Controller: client.NewController(ch, exchange, routes),
		net:        net,
		nets:       nets,
		run:        model,
		events:     events,
	}
}

func load(netClient *prisma.NetClient, runClient *prisma.RunClient, ch *amqp.Channel, exchange string, routes ...map[string]string) *Controller {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := runClient

	res, err := srv.List(ctx)
	if err != nil {
		panic(err)
	}

	fst := res[0]
	run, err := srv.Load(ctx, fst.ID)
	if err != nil {
		panic(err)
	}

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
			idxConstants := make(map[string]interface{})
			for _, c := range step.Action().Constants() {
				idxConstants[c.FieldID] = c.Value
			}
			if len(events[i].Fields) > 0 {
				fmt.Println("Adding fieldData")
				fieldData := make(map[string]interface{})
				for _, f := range step.Action().Event().Fields() {
					if val, ok := idxConstants[f.ID]; ok {
						if f.Type == "string" {
							fieldData[fLower(f.Name)] = val
						}
						if f.Type == "number" {
							floatVal, err := strconv.ParseFloat(val.(string), 64)
							if err != nil {
								panic(err)
							}
							fieldData[fLower(f.Name)] = floatVal
						}
						if f.Type == "boolean" {
							boolVal, err := strconv.ParseBool(val.(string))
							if err != nil {
								panic(err)
							}
							fieldData[fLower(f.Name)] = boolVal
						}
						continue
					}
					log.Fatalf("Missing fieldData for field %s", f.Name)
				}
				events[i].Data = fieldData
				if !events[i].IsValid() {
					panic("Invalid event")
				}

			}
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
	lnm := make(map[string]*labeled.Net)
	for _, n := range netClient.Nets {
		mn := marked.NewFromMap(n, netClient.InitialMarking)
		lnm[n.ID] = labeled.New(mn)
	}
	return NewController(events, ch, exchange, lnm, ln, run, rm)
}

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
				for _, t := range net.Transitions {
					for _, e := range step.Action().Device().Events() {
						if t.ID == e.ID {
							eventMap[e.Event().Name] = t
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

	devNameFromID := make(map[string]string)
	for devID, instanceID := range c.Routes {
		go func(devID, instanceID string) {
			fmt.Printf("Starting mock instance %s for device %s\n", instanceID, devID)
			srvConn, err := amqp.Dial(uri)
			if err != nil {
				panic(err)
			}
			srvCh, err := srvConn.Channel()
			if err != nil {
				panic(err)
			}
			h := make(map[string]labeled.Handler)
			for _, s := range c.run.Steps() {
				if s.Action().Device().ID == devID && s.Action().Device().Instances()[0].ID == instanceID {
					h[snakeCase(s.Action().Event().Name)] = makeHandler(s.Action().Event().Name)
				}
			}
			// maps event names to transitions
			for _, s := range c.run.Steps() {
				if s.Action().Device().ID == devID && s.Action().Device().Instances()[0].ID == instanceID {
					devNameFromID[devID] = s.Action().Device().Name
					for _, t := range s.Action().Event().Transitions() {
						devNet, found := deviceNetIndex[devID]
						if !found {
							panic("Device net not found")
						}
						if devNet.Transitions == nil {
							continue
						}
						for _, nt := range deviceNetIndex[devID].Transitions {
							if nt.ID == t.ID {
								eventMap[snakeCase(s.Action().Event().Name)] = nt
							}
						}
					}
				}
			}

			srv := server.New(deviceNetIndex[devID], srvCh, exchange, instanceID, eventMap, h)
			srv.Listen(ctx)
			<-done
		}(devID, instanceID)
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
		err = c.net.Fire(ev)
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
