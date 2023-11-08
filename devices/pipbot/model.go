package PipBot

import (
	"context"
	"errors"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	marlin "github.com/jt05610/petri/marlin/proto/v1"
	"strconv"
	"sync/atomic"
)

type PipetteTechnique int

const (
	Forward PipetteTechnique = 0
	Reverse PipetteTechnique = 1
)

type Fluid struct {
	Source   *Location `json:"source"`
	LastDest *Location `json:"lastDest"`
	Volume   float32   `json:"volume"`
}

func (f *Fluid) CanDispense(vol float32) bool {
	return f.Volume >= vol
}

func (f *Fluid) Dispense(vol float32) (*Fluid, error) {
	newVol := f.Volume - vol
	if newVol < 0 {
		return nil, errors.New("not enough fluid")
	}
	return &Fluid{
		Source: f.Source,
		Volume: newVol,
	}, nil
}

type State struct {
	Pipette *Fluid
	HasTip  bool
	*pipbot.Layout
	TipChannel    <-chan *pipbot.Position
	TipIndex      int
	needsPreRinse bool
	extruderPos   float32
}

type PipBot struct {
	marlin.MarlinServer
	*RedisClient
	state        *atomic.Pointer[State]
	transferCh   chan *TransferPlan
	transferring *atomic.Bool
	batch        []*StartTransferRequest
	batchRunning *atomic.Bool
}

func (d *PipBot) RunBatch(ctx context.Context) {
	d.batchRunning.Store(true)
	go func() {
		for _, req := range d.batch {
			_, err := d.StartTransfer(ctx, req)
			if err != nil {
				d.batchRunning.Store(false)
				return
			}
			_, err = d.FinishTransfer(ctx, &FinishTransferRequest{})
		}
		d.batchRunning.Store(false)
	}()
}

func (d *PipBot) Volumes() VolumeResponse {
	s := d.State()
	ret := make(VolumeResponse)
	for i, mat := range s.Layout.Matrices {
		ret[strconv.Itoa(i)] = mat.FluidLevelMap
	}
	return ret
}

func (d *PipBot) State() *State {
	return d.state.Load()
}

type StartTransferRequest struct {
	SourceGrid string  `json:"sourcegrid"`
	DestGrid   string  `json:"destgrid"`
	SourcePos  string  `json:"sourcepos"`
	DestPos    string  `json:"destpos"`
	Multi      float64 `json:"multi"`
	FlowRate   float64 `json:"flowrate"`
	Volume     float64 `json:"volume"`
	Technique  float64 `json:"technique"`
}

type StartTransferResponse struct {
	SourceGrid string  `json:"sourcegrid"`
	DestGrid   string  `json:"destgrid"`
	SourcePos  string  `json:"sourcepos"`
	DestPos    string  `json:"destpos"`
	Multi      float64 `json:"multi"`
	FlowRate   float64 `json:"flowrate"`
	Volume     float64 `json:"volume"`
	Technique  float64 `json:"technique"`
}

type FinishTransferRequest struct {
	Message string `json:"message"`
}

type FinishTransferResponse struct {
	Message string `json:"message"`
}
