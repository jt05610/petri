package autosampler

import (
	"context"
	"strconv"
)

type Autosampler struct {
	*Server
	finish context.CancelFunc
}

// vialFromGrid converts a grid position to a vial number. However, the rows increment backwards from J to A.
// Example J1 -> 1, J2 -> 2, A1 -> 90. The number of rows and columns are required to calculate the vial number.
func vialFromGrid(pos string, nRows, nCols int) (int32, error) {
	row := int(pos[0]-'A') + 1
	col, err := strconv.Atoi(pos[1:])
	if err != nil {
		return 0, err
	}
	return int32((nRows-row)*nCols + col), nil
}

type InjectRequest struct {
	InjectionVolume int    `json:"injectionvolume"`
	Position        string `json:"position"`
	AirCushion      int    `json:"aircushion"`
	ExcessVolume    int    `json:"excessvolume"`
	FlushVolume     int    `json:"flushvolume"`
	NeedleDepth     int    `json:"needledepth"`
}

func (r *InjectRequest) Requests() ([]*Request, error) {
	return InjectionSettings(r)
}

type InjectResponse struct {
	InjectionVolume int    `json:"injectionvolume"`
	Position        string `json:"position"`
	AirCushion      int    `json:"aircushion"`
	ExcessVolume    int    `json:"excessvolume"`
	FlushVolume     int    `json:"flushvolume"`
	NeedleDepth     int    `json:"needledepth"`
}

type InjectedRequest struct {
}

type InjectedResponse struct {
}

type WaitForReadyRequest struct {
}

type WaitForReadyResponse struct {
}
