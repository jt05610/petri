package autosampler

import (
	"context"
	"errors"
	autosampler "github.com/jt05610/petri/devices/autosampler/proto"
	"io"
	"strconv"
)

// vialFromGrid converts a grid position to a vial number. Example A1 -> 1, A2 -> 2. The number of rows and columns
// are required to calculate the vial number.
func vialFromGrid(pos string, nRows, nCols int) (int32, error) {
	row := int(pos[0] - 'A')
	if row >= nRows-1 {
		return 0, errors.New("row out of range")
	}
	col, err := strconv.Atoi(pos[1:])
	if err != nil {
		return 0, err
	}
	return int32(row*nCols + col), nil
}

func makeRequest(req *InjectRequest) (*autosampler.InjectRequest, error) {
	vial, err := vialFromGrid(req.Position, 10, 10)
	if err != nil {
		return nil, err
	}
	return &autosampler.InjectRequest{
		InjectionVolume: int32(req.InjectionVolume),
		Vial:            vial,
		AirCushion:      int32(req.AirCushion),
		FlushVolume:     int32(req.FlushVolume),
		ExcessVolume:    int32(req.ExcessVolume),
		NeedleDepth:     int32(req.NeedleDepth),
	}, nil
}

func (d *Autosampler) doInjection(ctx context.Context, client autosampler.Autosampler_InjectClient) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			resp, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			current := autosampler.InjectState(d.state.Load())
			if resp.State != current {
				d.stateChange <- resp.State
				d.state.Store(int32(resp.State))
			}
		}
	}
}

var ErrNotClearToInject = errors.New("not clear to inject")

func (d *Autosampler) Injected(ctx context.Context, _ *InjectedRequest) (*InjectedResponse, error) {
	err := d.waitFor(ctx, autosampler.InjectState_Injected)
	if err != nil {
		return nil, err
	}
	return &InjectedResponse{}, nil
}

func (d *Autosampler) waitFor(ctx context.Context, state autosampler.InjectState) error {
	if d.currentState() == state {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case upd := <-d.stateChange:
			if upd == state {
				return nil
			}
		}
	}
}

func (d *Autosampler) currentState() autosampler.InjectState {
	return autosampler.InjectState(d.state.Load())
}

func (d *Autosampler) Inject(ctx context.Context, req *InjectRequest) (*InjectResponse, error) {
	if d.currentState() != autosampler.InjectState_Idle {
		return nil, ErrNotClearToInject
	}
	r, err := makeRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Inject(ctx, r)
	if err != nil {
		return nil, err
	}
	d.state.Store(int32(autosampler.InjectState_Injecting))
	go func() {
		err := d.doInjection(ctx, resp)
		if err != nil {
			panic(err)
		}
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
	err := d.waitFor(ctx, autosampler.InjectState_Idle)
	if err != nil {
		return nil, err
	}
	return &WaitForReadyResponse{}, nil
}
