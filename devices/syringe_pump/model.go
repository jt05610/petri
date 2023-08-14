package main

import (
	modbus "github.com/jt05610/petri/proto/v1"
	"time"
)

type SyringePump struct {
	client    modbus.ModbusClient
	unitID    uint32
	increment time.Duration
	velCh     chan float64
}

type InitializeRequest struct {
}

type InitializeResponse struct {
}

type PumpRequest struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

type PumpResponse struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

type StopRequest struct {
}

type StopResponse struct {
}
