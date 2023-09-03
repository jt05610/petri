package main

type OrganicPump struct {
}

type InitializeRequest struct {
	SyringeDiameter float64 `json:"syringe_diameter"`
	SyringeVolume   float64 `json:"syringe_volume"`
	StepsPerMM      float64 `json:"steps_per_mm"`
	Rate            float64 `json:"rate"`
}

type InitializeResponse struct {
	SyringeDiameter float64 `json:"syringe_diameter"`
	SyringeVolume   float64 `json:"syringe_volume"`
	StepsPerMM      float64 `json:"steps_per_mm"`
	Rate            float64 `json:"rate"`
}

type PumpRequest struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

type PumpResponse struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

type PumpedRequest struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}

type PumpedResponse struct {
	Volume float64 `json:"volume"`
	Rate   float64 `json:"rate"`
}
