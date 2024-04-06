package main

import (
	"client/dbtl/proto/v1/dbtl"
	"context"
	"flag"
	"github.com/jt05610/petri/v1"
)

var fName string
var host string

func main() {
	flag.Parse()
	n := petri.LoadNet(fName)
	srv := NewService(n, nil)
	ctx, can := context.WithCancel(context.Background())
	defer can()
	go func() {
		err := srv.Listen(ctx)
		if err != nil {
			panic(err)
		}
	}()
	err := petri.Serve(ctx, host, n, dbtl.DbtlServiceServer(srv), dbtl.RegisterDbtlServiceServer)
	if err != nil {
		panic(err)
	}
}

func init() {
	flag.StringVar(&fName, "file", "", "path to the petri net file")
	flag.StringVar(&host, "host", "[::]:54445", "host to listen on")
}
