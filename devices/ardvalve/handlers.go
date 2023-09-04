package main

import (
	"context"
)

func (d *MixingValve) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	return &InitializeResponse{}, nil
}

func (d *MixingValve) Mix(ctx context.Context, req *MixRequest) (*MixResponse, error) {
	return &MixResponse{}, nil
}

func (d *MixingValve) Mixed(ctx context.Context, req *MixedRequest) (*MixedResponse, error) {
	return &MixedResponse{}, nil
}
