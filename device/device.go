package device

import (
	"github.com/jt05610/petri/labeled"
	"github.com/jt05610/petri/prisma"
)

type Message interface {
	RoutingKey() string
	Data() interface{}
	Validate() error
}

type Device struct {
	*labeled.Net
	*prisma.NetClient
}

func New(net *labeled.Net, client *prisma.NetClient) *Device {
	return &Device{
		Net:       net,
		NetClient: client,
	}
}
