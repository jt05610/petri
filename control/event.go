package control

import (
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma/db"
	"strings"
)

type Event struct {
	*labeled.Event
	Topic string
	From  string
}

func NewEvent(from string, fields []db.FieldModel, event *db.EventModel) *Event {
	ef := make([]*labeled.Field, len(fields))
	for i, f := range fields {
		ef[i] = &labeled.Field{
			Name: f.Name,
			Type: labeled.FieldType(f.Type),
		}
	}
	return &Event{
		Event: &labeled.Event{
			Name:   event.Name,
			Data:   event.Data,
			Fields: ef,
		},
		From: from,
	}
}

func (e *Event) snakeCaseName() string {
	return strings.Replace(strings.ToLower(e.Name), " ", "_", -1)
}

func (e *Event) RoutingKey() string {
	return e.From + ".events." + e.snakeCaseName()
}
