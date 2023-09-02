package main

import proto "github.com/jt05610/petri/grbl/proto/v1"

type SixPortRheodyneValve struct {
	proto.GRBLClient
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
