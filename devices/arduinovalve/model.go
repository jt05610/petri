package main

import "io"

type MixingValve struct {
	txCh chan []byte
	rxCh <-chan io.Reader
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
