package rheogrande

import (
	"context"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"log"
	"time"
)

// determined experimentally
var (
	Load   float32 = -21
	Inject float32 = 0
	Rate   float32 = 1500
)

func (d *SixPortRheodyneValve) OpenA(ctx context.Context, req *OpenARequest) (*OpenAResponse, error) {
	log.Printf("open a with delay %f", req.Delay)
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

	return &OpenAResponse{}, nil
}

func (d *SixPortRheodyneValve) OpenB(ctx context.Context, req *OpenBRequest) (*OpenBResponse, error) {
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
	return &OpenBResponse{}, nil
}
