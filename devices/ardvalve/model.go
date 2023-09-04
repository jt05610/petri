package main

import "github.com/jt05610/petri/grbl"

type MixingValve struct {
	txCh   chan []byte
	rxCh   <-chan []byte
	status *grbl.Status
}

type InitializeRequest struct {
	Components string `json:"components"`
}

type InitializeResponse struct {
	Components string `json:"components"`
}

type MixRequest struct {
	Proportions string `json:"proportions"`
}

type MixResponse struct {
	Proportions string `json:"proportions"`
}

type MixedRequest struct {
	Proportions string `json:"proportions"`
}

type MixedResponse struct {
	Proportions string `json:"proportions"`
}
