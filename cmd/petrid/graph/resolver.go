package graph

import (
	"context"
	"github.com/jt05610/petri/amqp/client"
	"github.com/jt05610/petri/cmd/petrid/graph/model"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/prisma"
	"github.com/jt05610/petri/prisma/db"
	"time"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*prisma.SessionClient
	*prisma.RunClient
	*prisma.NetClient
	*client.Controller
	eventCh <-chan *control.Event
}

func NewResolver(cl *db.PrismaClient, controller *client.Controller) *Resolver {
	r := &Resolver{
		SessionClient: &prisma.SessionClient{PrismaClient: cl},
		RunClient:     &prisma.RunClient{PrismaClient: cl},
		NetClient:     &prisma.NetClient{PrismaClient: cl},
		Controller:    controller,
		eventCh:       controller.Data(),
	}
	return r
}

func TimeStamp() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (r *Resolver) eventHistory(ctx context.Context, sessionID string) ([]*model.Event, error) {
	data, err := r.SessionData(ctx, sessionID)
	dModel := make([]*model.Event, len(data))
	for i, d := range data {
		val, ok := d.Value()
		if !ok {
			return nil, err
		}
		out, err := model.UnmarshalJSON(val)
		if err != nil {
			return nil, err
		}
		dModel[i] = &model.Event{
			Name: d.Event().Name,
			Data: out,
		}
	}
	return dModel, nil
}
