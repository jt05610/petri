package main

import (
	"context"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"time"
)

func (d *TwoPositionThreeWayValve) OpenA(ctx context.Context, req *OpenARequest) (*OpenAResponse, error) {
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
	}
	_, err := d.SpindleOff(ctx, &proto.SpindleOffRequest{})
	if err != nil {
		return nil, err
	}
	return &OpenAResponse{}, nil
}

func (d *TwoPositionThreeWayValve) OpenB(ctx context.Context, req *OpenBRequest) (*OpenBResponse, error) {
	if req.Delay > 0 {
		time.Sleep(time.Duration(req.Delay) * time.Millisecond)
	}
	_, err := d.SpindleOn(ctx, &proto.SpindleOnRequest{})
	if err != nil {
		return nil, err
	}
	return &OpenBResponse{}, nil
}
