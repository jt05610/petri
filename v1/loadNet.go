package petri

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	"google.golang.org/grpc"
	"net"
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

type registerFunc[T any] func(registrar grpc.ServiceRegistrar, srv T)

func Serve[T any](ctx context.Context, host string, n *petri.Net, srv T, reg registerFunc[T]) error {
	ctx, can := context.WithCancel(context.Background())
	defer can()
	server := grpc.NewServer()
	reg(server, srv)
	lis, err := net.Listen("tcp", host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("{\"%s\": \"%s\"}", n.Name, lis.Addr().String())
	return server.Serve(lis)
}
