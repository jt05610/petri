package camera

type Camera struct {
}

type CaptureRequest struct {
	Duration float64 `json:"duration"`
	Interval float64 `json:"interval"`
}

type CaptureResponse struct {
	Duration float64 `json:"duration"`
	Interval float64 `json:"interval"`
}

type ImagesCapturedRequest struct {
	Url string `json:"url"`
}

type ImagesCapturedResponse struct {
	Url string `json:"url"`
}
