package yaml

import (
	"context"
	"github.com/jt05610/petri"
	pf "github.com/jt05610/petri/petrifile"
	"github.com/jt05610/petri/petrifile/v1"
	"gopkg.in/yaml.v3"
	"io"
)

var _ pf.Service = (*Service)(nil)

type Service struct {
}

func (s *Service) Load(_ context.Context, r io.Reader) (*petri.Net, error) {
	var f petrifile.Petrifile
	err := yaml.NewDecoder(r).Decode(&f)
	if err != nil {
		return nil, err
	}
	return f.Net(), nil
}

func (s *Service) Save(_ context.Context, w io.Writer, n *petri.Net) error {
	return nil
}

func (s *Service) Version() pf.Version {
	return pf.V1
}
