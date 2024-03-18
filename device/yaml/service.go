package yaml

import (
	"context"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/petrifile"
	py "github.com/jt05610/petri/petrifile/v1/yaml"
	"gopkg.in/yaml.v3"
	"io"
)

var _ device.Service = (*Service)(nil)

type Service struct {
	petrifile.Service
	*builder.Builder
}

type DeviceFile struct {
	Petri    petrifile.Version `yaml:"petri"`
	Protocol string            `yaml:"protocol"`
	Net      string            `yaml:"net"`
	Remotes  map[string]string `yaml:"remotes"`
}

func (s *Service) Load(ctx context.Context, r io.Reader) (*device.Device, error) {
	var f DeviceFile
	err := yaml.NewDecoder(r).Decode(&f)
	if err != nil {
		return nil, err
	}
	net, err := s.Builder.Build(ctx, f.Net)
	if err != nil {
		return nil, err
	}
	return &device.Device{Net: net, Remotes: f.Remotes}, nil
}

func (s *Service) Save(ctx context.Context, w io.Writer, d *device.Device) error {
	//TODO implement me
	panic("implement me")
}

func NewService(dirs ...string) device.Service {
	bld := builder.NewBuilder(nil, dirs...)
	r := py.NewService(bld)
	bld = bld.WithService("yaml", r).WithService("yml", r)
	return &Service{Service: r, Builder: bld}
}
