package graph

import (
	"github.com/jt05610/petri/amqp/client"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/prisma"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*prisma.SessionClient
	*client.Controller
	eventCh <-chan *control.Event
}

func NewResolver(sessionClient prisma.SessionClient, controller *client.Controller) *Resolver {
	return &Resolver{
		SessionClient: &prisma.SessionClient{},
		Controller:    controller,
		eventCh:       controller.Data(),
	}
}
