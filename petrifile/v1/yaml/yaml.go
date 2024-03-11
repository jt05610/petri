package yaml

import (
	"context"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	pf "github.com/jt05610/petri/petrifile"
	"github.com/jt05610/petri/petrifile/v1"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

var _ pf.Service = (*Service)(nil)

type Service struct {
	bld *builder.Builder
}

func (s *Service) FileVersion(f string) (pf.Version, error) {
	var v struct {
		Petri pf.Version `yaml:"petri"`
	}

	file, err := os.Open(f)
	if err != nil {
		return "Unknown", err
	}
	err = yaml.NewDecoder(file).Decode(&v)
	if err != nil {
		return "Unknown", err
	}
	return v.Petri, nil
}

func (s *Service) Load(_ context.Context, r io.Reader) (*petri.Net, error) {
	var f petrifile.Petrifile
	err := yaml.NewDecoder(r).Decode(&f)
	if err != nil {
		return nil, err
	}
	f.Builder = s.bld
	return f.Net(), nil
}

func (s *Service) Save(_ context.Context, w io.Writer, n *petri.Net) error {
	return nil
}

func (s *Service) Version() pf.Version {
	return pf.V1
}

func NewService(b *builder.Builder) *Service {
	return &Service{bld: b}
}