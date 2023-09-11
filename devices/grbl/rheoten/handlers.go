package rheoten

import (
	"context"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"time"
)

func (d *TenPortRheodyneValve) OpenA(ctx context.Context, req *OpenARequest) (*OpenAResponse, error) {
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
	}
	_, err := d.CoolantOff(ctx, &proto.CoolantOffRequest{})
	if err != nil {
		return nil, err
	}
	_, err = d.MistOn(ctx, &proto.MistOnRequest{})
	if err != nil {
		return nil, err
	}

	return &OpenAResponse{}, nil
}

func (d *TenPortRheodyneValve) OpenB(ctx context.Context, req *OpenBRequest) (*OpenBResponse, error) {
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
	}
	_, err := d.CoolantOff(ctx, &proto.CoolantOffRequest{})
	if err != nil {
		return nil, err
	}
	_, err = d.FloodOn(ctx, &proto.FloodOnRequest{})
	if err != nil {
		return nil, err
	}
	return &OpenBResponse{}, nil
}
