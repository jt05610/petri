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
	f := petrifile.New(n)
	return yaml.NewEncoder(w).Encode(f)
}

func (s *Service) Version() pf.Version {
	return pf.V1
}

func NewService(b *builder.Builder) pf.Service {
	return &Service{bld: b}
}

func Load(rdr io.Reader) *petri.Net {
	srv := &Service{}
	net, err := srv.Load(context.Background(), rdr)
	if err != nil {
		panic(err)
	}
	return net
}

func LoadFile(f string) *petri.Net {
	file, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return Load(file)
}
