package camera_test

import (
	"context"
	camera "github.com/jt05610/petri/devices/gocam"
	"testing"
)

func TestCamera_Capture(t *testing.T) {
	cam := camera.NewCamera()
	_, err := cam.Capture(context.Background(), &camera.CaptureRequest{
		Duration: 5,
		Interval: 100,
	})
	if err != nil {
		t.Error(err)
	}
}
