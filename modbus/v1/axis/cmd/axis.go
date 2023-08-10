package cmd

import (
	"context"
	"core/axis/v1/axis"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	mbAxis "modbus/v1/axis"
	"modbus/v1/modbus"
	"net"
	"os"
)

func Execute(ctx context.Context, addr, mbAddr, certFn, keyFn string) {
	flag.Parse()
	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	log.Printf("Axis server listening on https://%s/\n", addr)
	db, err := gorm.Open(sqlite.Open("axis.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&mbAxis.ModbusAxis{}, &axis.Calibration{}, &modbus.InputRegister{}, &modbus.HoldingRegister{}, &modbus.Coil{}, &modbus.DiscreteInput{})
	server := grpc.NewServer()
	store := mbAxis.NewStore(db)
	mbAxis.RegisterStoreServer(server, store)
	axs, err := store.ListAxis(context.Background(), &axis.Empty{})
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	caCert, err := os.ReadFile(certFn)
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(caCert)
	conn, err := grpc.Dial(
		mbAddr,
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
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()
	client := modbus.NewModbusClient(conn)
	axis.RegisterDeviceServer(server, mbAxis.NewServer(client, axs.Devices))
	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()
	if err := server.Serve(
		tls.NewListener(
			lis,
			&tls.Config{
				Certificates:     []tls.Certificate{cert},
				CurvePreferences: []tls.CurveID{tls.CurveP256},
				MinVersion:       tls.VersionTLS12,
			},
		),
	); err != nil {
		log.Fatal(err)
	}
	log.Println("Axis server stopped")
}
