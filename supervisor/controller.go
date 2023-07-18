package supervisor

import (
	"context"
	"pnet/net"
)

type PetriNetController struct {
	net        *net.PetriNet
	eventQueue chan struct{}
	cancel     context.CancelFunc
}

func NewController(net *net.PetriNet) *PetriNetController {
	return &PetriNetController{
		net:        net,
		eventQueue: make(chan struct{}),
	}
}

func (c *PetriNetController) Start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-c.eventQueue:
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *PetriNetController) Stop() {
	c.cancel()
}
