package control

import (
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma/db"
	"strings"
)

type Command struct {
	*labeled.Event
	Topic string
	To    string
}

func NewCommand(to string, fields []db.FieldModel, event *db.EventModel, data interface{}) *Command {
	ef := make([]*labeled.Field, len(fields))
	for i, f := range fields {
		ef[i] = &labeled.Field{
			Name: f.Name,
			Type: labeled.FieldType(f.Type),
		}
	}
	return &Command{
		Event: &labeled.Event{
			Name:   event.Name,
			Fields: ef,
			Data:   data,
		},
		To: to,
	}
}

func (c *Command) snakeCaseName() string {
	return strings.Replace(strings.ToLower(c.Name), " ", "_", -1)
}

func (c *Command) RoutingKey() string {
	return c.To + ".commands." + c.snakeCaseName()
}
