package main

import (
	"context"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"time"
)

// determined experimentally
var (
	Load   float32 = 0.0
	Inject float32 = -10.9
	Rate   float32 = 400
)

func (d *SixPortRheodyneValve) OpenA(ctx context.Context, req *OpenARequest) (*OpenAResponse, error) {
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
	}
	_, err := d.Move(ctx, &proto.MoveRequest{
		Z:     &Load,
		Speed: &Rate,
	})
	if err != nil {
		return nil, err
	}

	return &OpenAResponse{}, nil
}

func (d *SixPortRheodyneValve) OpenB(ctx context.Context, req *OpenBRequest) (*OpenBResponse, error) {
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
	}
	_, err := d.Move(ctx, &proto.MoveRequest{
		Z:     &Inject,
		Speed: &Rate,
	})
	if err != nil {
		return nil, err
	}
	return &OpenBResponse{}, nil
}
