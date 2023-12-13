package yaml

import (
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/db"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/prisma/db"
	"gopkg.in/yaml.v3"
	"io"
)

type Service struct {
}

func (s *Service) Load(r io.Reader) (*db.DeviceModel, error) {
	dec := yaml.NewDecoder(r)
	var model db.DeviceModel
	return &model, dec.Decode(&model)
}

func (s *Service) Flush(w io.Writer, model *db.DeviceModel) error {
	enc := yaml.NewEncoder(w)
	return enc.Encode(model)
}

func (s *Service) ToNet(model *db.DeviceModel, handlers control.Handlers) (*device.Device, error) {
	return prisma.ConvertDevice(model, handlers)
}
