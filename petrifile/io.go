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
	FileVersion(f string) (Version, error)
}

type Version string

const (
	V1      Version = "v1"
	Unknown Version = "Unknown"
)
