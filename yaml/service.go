package yaml

import (
	"github.com/jt05610/petri/device"
	"io"
)

type DeviceService struct {
	Filename string
}

func (s *DeviceService) Load(r io.Reader) (*device.Device, error) {
	panic("not implemented")
}

func (s *DeviceService) Flush(w io.Writer, model *device.Device) error {
	panic("not implemented")
}
