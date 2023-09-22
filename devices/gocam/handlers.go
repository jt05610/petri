package camera

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math"
	"net/http"
)

const camUrl = "http://127.0.0.1:8000/record/"

type Result struct {
	Message string
}

func (d *Camera) Capture(ctx context.Context, req *CaptureRequest) (*CaptureResponse, error) {
	id := uuid.New()
	body := struct {
		URL       string `json:"url,omitempty"`
		Name      string `json:"name,omitempty"`
		Duration  int    `json:"duration,omitempty"`
		Frequency int    `json:"frequency,omitempty"`
	}{
		URL:       camUrl,
		Name:      id.String(),
		Duration:  int(math.Round(req.Duration)),
		Frequency: int(math.Round(1000 / req.Interval)), // converting from ms to Hz
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(camUrl, "application/json", buf)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	result := new(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return nil, err
	}
	fmt.Printf(res.Status)
	fmt.Println(result)
	return &CaptureResponse{}, nil
}

func (d *Camera) ImagesCaptured(ctx context.Context, req *ImagesCapturedRequest) (*ImagesCapturedResponse, error) {
	return &ImagesCapturedResponse{}, nil
}
