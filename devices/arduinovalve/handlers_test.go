package main

import (
	"context"
	"go.uber.org/zap"
	"testing"
	"time"
)

func mix(ctx context.Context, t *testing.T, d *MixingValve, props string, period uint64) {
	ctx, timeout := context.WithTimeout(ctx, time.Duration(100)*time.Millisecond)
	defer timeout()
	_, err := d.Mix(context.Background(), &MixRequest{
		Proportions: props,
		Period:      period,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Mixed(context.Background(), &MixedRequest{})
	if err != nil {
		t.Fatal(err)
	}
}

func runRequests(ctx context.Context, t *testing.T, d *MixingValve) {
	mix(ctx, t, d, "0,1,0,0,1,0,0", 100)
	mix(ctx, t, d, "0,0,0,0,0,0,1", 100)
}

func TestMixingValve_Mix(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			t.Fatal(err)
		}
	}()
	cfg, err := LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := new(MixingValve)
	go Run(ctx, d, logger, cfg)
	<-time.After(time.Duration(2) * time.Second)
	for i := 0; i < 10000; i++ {
		runRequests(ctx, t, d)
	}
}
