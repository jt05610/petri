package control

import (
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/marked"
)

type Device struct {
	ID string
	*marked.Net
	// Handlers are called when a transition is fired
	handlers map[string]labeled.Handler
	mu       chan struct{}
}

type DeviceMap map[string]*Device

type Controller struct {
	devices DeviceMap
}
