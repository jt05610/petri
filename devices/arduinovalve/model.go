package main

type MixingValve struct {
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
