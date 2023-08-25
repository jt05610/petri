package main

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jt05610/petri/amqp/client"
	"github.com/jt05610/petri/cmd/petrid/graph"
	"github.com/jt05610/petri/cmd/petrid/graph/generated"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
)

func main() {
	dbClient := db.NewClient()
	if err := dbClient.Connect(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := dbClient.Disconnect(); err != nil {
			log.Fatal(err)
		}
	}()
	uri, found := os.LookupEnv("RABBITMQ_URI")
	if !found {
		panic(errors.New("RABBITMQ_URI not set"))
	}
	conn, err := amqp.Dial(uri)
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	exchange, found := os.LookupEnv("AMQP_EXCHANGE")
	if !found {
		panic(errors.New("AMQP_EXCHANGE not set"))
	}
	controller := client.NewController(ch, exchange)
	defer controller.Close()
	controller.Listen(context.Background())
	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: &graph.Resolver{
					SessionClient: &prisma.SessionClient{PrismaClient: dbClient},
					RunClient:     &prisma.RunClient{PrismaClient: dbClient},
					Controller:    controller,
				},
			},
		),
	)
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
		// send this panic somewhere
		log.Print(err)
		return errors.New("user message on panic")
	})

	http.Handle("/", srv)
	http.Handle("/playground", playground.Handler("Session", "/api/"))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
