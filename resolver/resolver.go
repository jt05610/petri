package resolver

//go:generate go run github.com/99designs/gqlgen generate

import (
	"context"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/amqp/client"
	"github.com/jt05610/petri/control"
)

type Resolver struct {
	*client.Controller
	tokenSchema  petri.Service[*petri.TokenSchema, *petri.TokenSchemaInput, *petri.TokenFilter, *petri.TokenUpdate]
	places       petri.Service[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate]
	transitions  petri.Service[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate]
	eventSchema  petri.Service[*petri.EventSchema, *petri.EventInput, *petri.EventFilter, *petri.EventUpdate]
	arcs         petri.Service[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate]
	nets         petri.Service[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate]
	dataCh       chan *control.Event
	seenEvents   map[string]int
	recordCtx    context.Context
	recordCancel context.CancelFunc
	runCtx       context.Context
	runCancel    context.CancelFunc
}

func NewResolver(
	tokenSchema petri.Service[*petri.TokenSchema, *petri.TokenSchemaInput, *petri.TokenFilter, *petri.TokenUpdate],
	places petri.Service[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate],
	transitions petri.Service[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate],
	arcs petri.Service[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate],
	nets petri.Service[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate],
	eventSchema petri.Service[*petri.EventSchema, *petri.EventInput, *petri.EventFilter, *petri.EventUpdate],
) *Resolver {
	return &Resolver{
		tokenSchema: tokenSchema,
		places:      places,
		transitions: transitions,
		arcs:        arcs,
		nets:        nets,
		eventSchema: eventSchema,
		dataCh:      make(chan *control.Event),
		seenEvents:  make(map[string]int),
	}
}
