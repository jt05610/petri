package autosampler

import (
	"context"
	"errors"
)

var ErrNotClearToInject = errors.New("not clear to inject")

func (d *Autosampler) Injected(ctx context.Context, _ *InjectedRequest) (*InjectedResponse, error) {
	err := d.waitFor(ctx, Finishing)
	if err != nil {
		return nil, err
	}
	return &InjectedResponse{}, nil
}

func (d *Autosampler) waitFor(ctx context.Context, state State) error {
	if d.currentState() == state {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if d.currentState() == state {
				return nil
			}
		}
	}
}

func (d *Autosampler) currentState() State {
	return State(d.state.Load())
}

func (d *Autosampler) Inject(ctx context.Context, req *InjectRequest) (*InjectResponse, error) {
	if d.currentState() != Idle {
		return nil, ErrNotClearToInject
	}
	d.state.Store(uint32(Injecting))
	hbCtx, can := context.WithCancel(ctx)
	d.finish = can
	go func() {
		err := d.startInjection(hbCtx, req)
		if err != nil {
			panic(err)
		}
		d.RunHeartbeat(hbCtx)
	}()
	return &InjectResponse{
		Position:        req.Position,
		InjectionVolume: req.InjectionVolume,
		AirCushion:      req.AirCushion,
		FlushVolume:     req.FlushVolume,
		ExcessVolume:    req.ExcessVolume,
		NeedleDepth:     req.NeedleDepth,
	}, nil
}

func (d *Autosampler) WaitForReady(ctx context.Context, _ *WaitForReadyRequest) (*WaitForReadyResponse, error) {
	defer d.finish()
	err := d.waitFor(ctx, Idle)
	if err != nil {
		return nil, err
	}
	return &WaitForReadyResponse{}, nil
}
