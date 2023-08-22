package main

import (
	"context"
)

func (d *TwoPositionThreeWayValve) OpenA(_ context.Context, _ *OpenARequest) (*OpenAResponse, error) {
	err := d.do(OpenA)
	if err != nil {
		return nil, err
	}
	return &OpenAResponse{}, nil
}

func (d *TwoPositionThreeWayValve) OpenB(_ context.Context, _ *OpenBRequest) (*OpenBResponse, error) {
	err := d.do(OpenB)
	if err != nil {
		return nil, err
	}
	return &OpenBResponse{}, nil
}
