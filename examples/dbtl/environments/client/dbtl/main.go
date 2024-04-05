package main

import (
	"client/dbtl/proto/v1/dbtl"
	"context"
	"flag"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	"google.golang.org/grpc"
	"net"
	"os"
	"path/filepath"
)

var fName string
var host string

func main() {
	flag.Parse()
	n := loadNet(fName)
	srv := NewService(n, nil)
	ctx, can := context.WithCancel(context.Background())
	defer can()
	go func() {
		err := srv.Listen(ctx)
		if err != nil {
			panic(err)
		}
	}()

	server := grpc.NewServer()

	dbtl.RegisterDbtlServiceServer(server, srv)
	lis, err := net.Listen("tcp", host)
	if err != nil {
		panic(err)
	}
	err = server.Serve(lis)
	if err != nil {
		panic(err)
	}
}

func loadNet(fName string) *petri.Net {
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

func init() {
	flag.StringVar(&fName, "file", "", "path to the petri net file")
	flag.StringVar(&host, "host", "[:]:0", "host to listen on")
}
