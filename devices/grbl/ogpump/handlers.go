package main

import (
	"context"
)

func (d *OrganicPump) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	return &InitializeResponse{}, nil
}

func (d *OrganicPump) Pump(ctx context.Context, req *PumpRequest) (*PumpResponse, error) {
	return &PumpResponse{}, nil
}

func (d *OrganicPump) Pumped(ctx context.Context, req *PumpedRequest) (*PumpedResponse, error) {
	return &PumpedResponse{}, nil
}
