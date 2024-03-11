package builder

import (
	"context"
	"errors"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/petrifile"
	"os"
	"path/filepath"
	"strings"
)

type Builder struct {
	SearchDirs []string
	seen       map[string]*petri.Net
	services   map[string]map[petrifile.Version]petrifile.Service
}

func NewBuilder(services map[string]map[petrifile.Version]petrifile.Service, dirs ...string) *Builder {
	if dirs == nil {
		dirs = []string{"."}
	}
	return &Builder{
		SearchDirs: dirs,
		seen:       make(map[string]*petri.Net),
		services:   services,
	}
}

func (b *Builder) WithService(s string, srv petrifile.Service) *Builder {
	if b.services == nil {
		b.services = make(map[string]map[petrifile.Version]petrifile.Service)
	}
	if _, ok := b.services[s]; !ok {
		b.services[s] = make(map[petrifile.Version]petrifile.Service)
	}
	b.services[s][srv.Version()] = srv
	return b
}

func (b *Builder) WithSearchDirs(dirs ...string) *Builder {
	b.SearchDirs = append(b.SearchDirs, dirs...)
	return b
}

func (b *Builder) service(f string) petrifile.Service {
	fileExt := filepath.Ext(f)
	if fileExt == "" {
		fileExt = ".yaml"
	}
	noDot := strings.Trim(fileExt, ".")
	if s, ok := b.services[noDot]; ok {
		var ver petrifile.Version
		var err error
		for _, v := range s {
			for _, d := range b.SearchDirs {
				ver, err = v.FileVersion(filepath.Join(d, f))
				if err != nil {
					if !errors.Is(err, os.ErrNotExist) {
						panic(err)
					}
				}
				if ver != petrifile.Unknown {
					break
				}
			}
			if ver != petrifile.Unknown {
				break
			}
		}
		srv, ok := s[ver]
		if !ok {
			panic("unknown version: " + ver)
		}
		return srv
	}
	return nil

}

func (b *Builder) Build(ctx context.Context, f string) (*petri.Net, error) {
	// find the file in the search dirs
	for _, dir := range b.SearchDirs {
		// if the file is found, load it
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.Name() == f {
				file, err := os.Open(filepath.Join(dir, f))
				if err != nil {
					return nil, err
				}
				srv := b.service(f)
				if srv == nil {
					return nil, os.ErrNotExist
				}
				n, err := srv.Load(ctx, file)
				if err != nil {
					return nil, err
				}
				_, seen := b.seen[f]
				if !seen {
					b.seen[f] = n
				}
				if n.Nets != nil {
					newNets := make([]*petri.Net, len(n.Nets))
					for i, net := range n.Nets {
						if subnet, seen := b.seen[net.Name]; seen {
							newNets[i] = subnet
						} else {
							newNet, err := b.Build(ctx, net.Name)
							if err != nil {
								return nil, err
							}
							b.seen[net.Name] = newNet
							newNets[i] = newNet
						}
					}
					n.Nets = newNets
				}
				return n, nil
			}
		}
	}
	return nil, os.ErrNotExist
}
