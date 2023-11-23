package control

import (
	"github.com/jt05610/petri/labeled"
	"strings"
)

type MarkingUpdate struct {
	DeviceID string
	Marking  Marking
}

type Marking map[string]int

func (m *Marking) JSON() map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range *m {
		ret[k] = v
	}
	return ret
}

type Event struct {
	*labeled.Event
	Topic   string
	From    string
	Marking Marking
}

func (e *Event) snakeCaseName() string {
	return strings.Replace(strings.ToLower(e.Name), " ", "_", -1)
}

func (e *Event) RoutingKey() string {
	return e.From + ".events." + e.snakeCaseName()
}
