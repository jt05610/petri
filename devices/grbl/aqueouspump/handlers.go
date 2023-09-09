package main

import (
	"context"
	"errors"
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"math"
)

func (d *AqueousPump) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	d.Rate = req.Rate
	d.SyringeDiameter = req.SyringeDiameter
	d.SyringeVolume = req.SyringeVolume
	d.StepsPerMM = req.StepsPerMM
	d.Pos = 0
	return &InitializeResponse{
		SyringeDiameter: d.SyringeDiameter,
		SyringeVolume:   d.SyringeVolume,
		StepsPerMM:      d.StepsPerMM,
		Rate:            d.Rate,
	}, nil
}

func (d *AqueousPump) VolToMM(vol float64) *float32 {
	ret := float32(1000 * vol / (math.Pow(d.SyringeDiameter/2, 2) * math.Pi))
	return &ret
}

func (d *AqueousPump) Pump(ctx context.Context, req *PumpRequest) (*PumpResponse, error) {
	dist := d.VolToMM(req.Volume)
	if dist == nil {
		return nil, errors.New("no volume")
	}
	mm := *dist
	newPos := mm + d.Pos
	if newPos > d.MaxPos || newPos < 0 {
		return nil, errors.New("new pos error")
	}

	_, err := d.Move(ctx, &proto.MoveRequest{
		Y:     &newPos,
		Speed: d.VolToMM(req.Rate),
	})
	if err != nil {
		return nil, err
	}
	d.Pos = newPos
	return &PumpResponse{
		Volume: req.Volume,
		Rate:   req.Rate,
	}, nil
}

func (d *AqueousPump) Pumped(ctx context.Context, req *PumpedRequest) (*PumpedResponse, error) {
	return &PumpedResponse{
		Volume: req.Volume,
		Rate:   req.Rate,
	}, nil
}
