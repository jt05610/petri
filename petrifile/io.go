package petrifile

import (
	"context"
	"github.com/jt05610/petri"
	"io"
)

type Service interface {
	Load(ctx context.Context, r io.Reader) (*petri.Net, error)
	Save(ctx context.Context, w io.Writer, n *petri.Net) error
	Version() Version
}

type Version string

const (
	V1 Version = "v1"
)
