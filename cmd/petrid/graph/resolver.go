package graph

import (
	"context"
	"errors"
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
	dataCh        chan *control.Event
	sessionEvents map[string][]*model.Event
	seenEvents    map[string]int
	recordCtx     context.Context
	recordCancel  context.CancelFunc
	runCtx        context.Context
	runCancel     context.CancelFunc
}

func NewResolver(cl *db.PrismaClient, controller *client.Controller) *Resolver {
	r := &Resolver{
		SessionClient: &prisma.SessionClient{PrismaClient: cl},
		RunClient:     &prisma.RunClient{PrismaClient: cl},
		NetClient:     &prisma.NetClient{PrismaClient: cl},
		Controller:    controller,
		sessionEvents: make(map[string][]*model.Event),
		runCtx:        context.Background(),
		recordCtx:     context.Background(),
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

func (r *Resolver) start(sessionId string) *model.Event {
	r.runCtx, r.runCancel = context.WithCancel(r.runCtx)
	ch := make(chan *control.Event)
	r.Controller.ChannelData(ch)
	go func() {
		for {
			select {
			case <-r.runCtx.Done():
				return
			case data := <-ch:
				r.sessionEvents[sessionId] = append(r.sessionEvents[sessionId], &model.Event{
					Name: data.Event.Name,
					Data: data.Data,
				})
				_, err := r.SessionClient.AddData(r.runCtx, sessionId, data)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
	r.Start(r.runCtx)
	return &model.Event{
		Name: "start",
		Data: nil,
	}
}

func (r *Resolver) stopRecording() {
	r.recordCancel()
	r.runCancel()
}

func (r *Resolver) newEvents(sessionID string) ([]*model.Event, error) {
	events, found := r.sessionEvents[sessionID]
	if !found {
		return nil, errors.New("session not found")
	}
	if len(events) <= r.seenEvents[sessionID] {
		return nil, nil
	}
	newEvents := events[r.seenEvents[sessionID]:]
	r.seenEvents[sessionID] = len(events)
	return newEvents, nil
}
