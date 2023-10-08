package main

import (
	"io"
	"sync/atomic"
)

type MixingValve struct {
	txCh   chan []byte
	rxCh   <-chan io.Reader
	period *atomic.Int32
	cts    *atomic.Bool
}

type MixRequest struct {
	Proportions string `json:"proportions"`
	Period      uint64 `json:"period"`
}

type MixResponse struct {
	Proportions string `json:"proportions"`
	Period      uint64 `json:"period"`
}

type MixedRequest struct {
	Proportions string `json:"proportions"`
}

type MixedResponse struct {
	Proportions string `json:"proportions"`
}
