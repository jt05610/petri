package control

import (
	"github.com/jt05610/petri/labeled"
	"strings"
)

type Command struct {
	*labeled.Event
	Topic string
	To    string
}

func (c *Command) snakeCaseName() string {
	return strings.Replace(strings.ToLower(c.Name), " ", "_", -1)
}

func (c *Command) RoutingKey() string {
	return c.To + ".commands." + c.snakeCaseName()
}
