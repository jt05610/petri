package main

import (
	"context"
)

func (d *TwoPositionThreeWayValve) OpenA(ctx context.Context, req *OpenARequest) (*OpenAResponse, error) {
	d.txCh <- []byte("M5\n")
	return &OpenAResponse{}, nil
}

func (d *TwoPositionThreeWayValve) OpenB(ctx context.Context, req *OpenBRequest) (*OpenBResponse, error) {
	//d.txCh <- []byte("M3\n")
	return &OpenBResponse{}, nil
}
