package main

import (
	"context"
	"github.com/jt05610/petri/codegen"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
)

func main() {
	devClient := &prisma.DeviceClient{PrismaClient: db.NewClient()}
	g := codegen.NewGenerator(devClient, &codegen.Params{
		Language:     "go",
		Port:         8080,
		OutDir:       "testResult",
		DeviceID:     "cll29i9oz002c0g5m8idsfw2n",
		RabbitMQURI:  "amqp://guest:guest@localhost:5672/",
		AMQPExchange: "topic_devices",
	})
	err := g.Generate(context.Background())
	if err != nil {
		panic(err)
	}
	// Output:
	// Generated code for device: cll29i9oz002c0g5m8idsfw2n
}
