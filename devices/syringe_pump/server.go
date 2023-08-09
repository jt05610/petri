package main

import (
	"context"
	"embed"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
	"github.com/jt05610/petri/prisma/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

//go:embed net.json
var netJSON embed.FS

//go:embed eventMap.json
var eventMapJSON embed.FS

type SyringePump struct {
}

func (v *SyringePump) Initialize(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *SyringePump) Pump(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *SyringePump) Stop(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *SyringePump) Handlers() control.Handlers {
	return control.Handlers{
		"initialize": v.Initialize,
		"pump":       v.Pump,
		"stop":       v.Stop,
	}
}

func convertFromDB(net *db.NetModel) *labeled.Net {
	places := make([]*petri.Place, len(net.Places()))
	transitions := make([]*petri.Transition, len(net.Transitions()))
	arcs := make([]*petri.Arc, len(net.Arcs()))
	nodeIndex := make(map[string]petri.Node)
	for i, place := range net.Places() {
		places[i] = &petri.Place{
			ID:    place.ID,
			Name:  place.Name,
			Bound: place.Bound,
		}
		nodeIndex[place.ID] = places[i]
	}
	for i, transition := range net.Transitions() {
		transitions[i] = &petri.Transition{
			ID:   transition.ID,
			Name: transition.Name,
		}
		nodeIndex[transition.ID] = transitions[i]
	}

	for i, arc := range net.Arcs() {
		if arc.FromPlace {
			arcs[i] = &petri.Arc{
				ID:   arc.ID,
				Src:  nodeIndex[arc.PlaceID],
				Dest: nodeIndex[arc.TransitionID],
			}
		} else {
			arcs[i] = &petri.Arc{
				ID:   arc.ID,
				Src:  nodeIndex[arc.TransitionID],
				Dest: nodeIndex[arc.PlaceID],
			}
		}
	}
	pNet := petri.New(places, transitions, arcs)
	markedNet := marked.New(pNet, net.InitialMarking)
	return labeled.New(markedNet)
}

func loadNet() *labeled.Net {
	df, err := netJSON.Open("net.json")
	failOnError(err, "Error opening net.json")
	var net db.NetModel
	decoder := json.NewDecoder(df)
	err = decoder.Decode(&net)
	failOnError(err, "Error decoding net.json")
	return convertFromDB(&net)
}

func loadEventMap() map[string]*petri.Transition {
	df, err := eventMapJSON.Open("eventMap.json")
	failOnError(err, "Error opening eventMap.json")
	var eventMap map[string]*petri.Transition
	decoder := json.NewDecoder(df)
	err = decoder.Decode(&eventMap)
	failOnError(err, "Error decoding eventMap.json")
	return eventMap
}

func main() {
	logger, err := zap.NewProduction()
	failOnError(err, "Error creating logger")
	err = godotenv.Load()
	failOnError(err, "Error loading .env file")

	logger.Info("Starting 🐰 server")
	// Setup rabbitmq channel
	uri, ok := os.LookupEnv("RABBITMQ_URI")
	if !ok {
		logger.Fatal("RABBITMQ_URI not set")
	}
	exchange, ok := os.LookupEnv("AMQP_EXCHANGE")
	if !ok {
		logger.Fatal("AMQP_EXCHANGE not set")
	}
	deviceID, ok := os.LookupEnv("DEVICE_ID")
	if !ok {
		logger.Fatal("DEVICE_ID not set")
	}
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")
	logger.Info("Connected to RabbitMQ", zap.String("uri", uri))
	defer func() {
		err := conn.Close()
		failOnError(err, "Failed to close connection")
	}()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	logger.Info("Opened channel")
	defer func() {
		err := ch.Close()
		failOnError(err, "Failed to close channel")
	}()
	pNet := loadNet()
	eventMap := loadEventMap()
	v := &SyringePump{}
	srv := server.New(pNet, ch, exchange, deviceID, eventMap, v.Handlers())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c // Wait for SIGINT
		cancel()
	}()
	srv.Listen(ctx)
	<-ctx.Done()
	logger.Info("Shutting down 🐰 server")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
