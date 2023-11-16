package autosampler

import (
	autosampler "github.com/jt05610/petri/devices/autosampler/proto"
	"sync/atomic"
)

type Autosampler struct {
	client      autosampler.AutosamplerClient
	state       *atomic.Int32
	stateChange chan autosampler.InjectState
}

var (
	_ Command = (*InjectRequest)(nil)
)

type InjectRequest struct {
	InjectionVolume float64 `json:"injectionvolume"`
	Position        string  `json:"position"`
	AirCushion      float64 `json:"aircushion"`
	ExcessVolume    float64 `json:"excessvolume"`
	FlushVolume     float64 `json:"flushvolume"`
	NeedleDepth     float64 `json:"needledepth"`
}

func (r *InjectRequest) Bytes() []byte {
	//TODO implement me
	panic("implement me")
}

type InjectResponse struct {
	InjectionVolume float64 `json:"injectionvolume"`
	Position        string  `json:"position"`
	AirCushion      float64 `json:"aircushion"`
	ExcessVolume    float64 `json:"excessvolume"`
	FlushVolume     float64 `json:"flushvolume"`
	NeedleDepth     float64 `json:"needledepth"`
}

type InjectedRequest struct {
}

type InjectedResponse struct {
}

type WaitForReadyRequest struct {
}

type WaitForReadyResponse struct {
}
