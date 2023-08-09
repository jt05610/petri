package main

import (
	"context"
	ax "core/axis/v1/axis"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"modbus/v1/axis"
	"os"
)

var axisAddr, caCertFn string

func init() {
	flag.StringVar(&axisAddr, "axisAddr", "localhost:55056", "address to listen on")
	flag.StringVar(&caCertFn, "caCert", "cert.pem", "CA certificate file")
}
func main() {
	flag.Parse()
	certPool := x509.NewCertPool()
	caCert, err := os.ReadFile(caCertFn)
	if err != nil {
		panic(err)
	}
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		panic("failed to append CA certificate")
	}
	conn, err := grpc.Dial(
		axisAddr,
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					CurvePreferences: []tls.CurveID{tls.CurveP256},
					MinVersion:       tls.VersionTLS12,
					RootCAs:          certPool,
				},
			),
		),
	)
	if err != nil {
		panic(err)
	}
	client := axis.NewStoreClient(conn)
	res, err := client.ListAxis(context.Background(), &ax.Empty{})
	if err != nil {
		panic(err)
	}
	for _, a := range res.Devices {
		println(a.Axis.Id)
	}
}
