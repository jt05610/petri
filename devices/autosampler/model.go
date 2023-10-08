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

type InjectRequest struct {
	InjectionVolume float64 `json:"injectionvolume"`
	Position        string  `json:"position"`
	AirCushion      float64 `json:"aircushion"`
	ExcessVolume    float64 `json:"excessvolume"`
	FlushVolume     float64 `json:"flushvolume"`
	NeedleDepth     float64 `json:"needledepth"`
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
