package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/jt05610/petri/codegen"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	devClient := &prisma.DeviceClient{PrismaClient: db.NewClient()}
	g := codegen.NewGenerator(devClient, &codegen.Params{
		Language:     "go",
		Port:         8080,
		OutDir:       "testResult",
		DeviceID:     "cll29i9p5002g0g5m2cw2jk6c",
		RabbitMQURI:  "amqp://guest:guest@localhost:5672/",
		AMQPExchange: "topic_devices",
	})
	authorID, found := os.LookupEnv("AUTHOR_ID")
	if !found {
		panic("AUTHOR_ID not found")
	}
	ctx := context.WithValue(context.Background(), "authorID", authorID)
	err = g.Generate(ctx)
	if err != nil {
		panic(err)
	}

}
