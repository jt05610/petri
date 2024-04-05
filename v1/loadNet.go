package petri

import (
	"context"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	"os"
	"path/filepath"
)

func LoadNet(fName string) *petri.Net {
	parDir := filepath.Dir(fName)
	bld := builder.NewBuilder(nil, ".", parDir)
	r := yaml.NewService(bld)
	bld = bld.WithService("yaml", r)
	bld = bld.WithService("yml", r)
	f, err := os.Open(fName)
	if err != nil {
		panic(err)
	}

	net, err := r.Load(context.Background(), f)
	if err != nil {
		panic(err)
	}
	return net
}
