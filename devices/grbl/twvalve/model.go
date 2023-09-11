package twvalve

import proto "github.com/jt05610/petri/grbl/proto/v1"

type TwoPositionThreeWayValve struct {
	proto.GRBLServer
}

type OpenARequest struct {
	Delay float64 `json:"delay"`
}

type OpenAResponse struct {
	Delay float64 `json:"delay"`
}

type OpenBRequest struct {
	Delay float64 `json:"delay"`
}

type OpenBResponse struct {
	Delay float64 `json:"delay"`
}
