package cmd

import (
	"context"
	"core/sensor/v1/sensor"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"modbus/v1/modbus"
	mbSensor "modbus/v1/sensor"
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
	log.Printf("Sensor server listening on https://%s/\n", addr)
	db, err := gorm.Open(sqlite.Open("sensor.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&mbSensor.ModbusSensor{})
	server := grpc.NewServer()
	store := mbSensor.NewStore(db)
	mbSensor.RegisterStoreServer(server, store)
	sensors, err := store.ListSensor(context.Background(), &sensor.Empty{})
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
	sensor.RegisterDeviceServer(server, mbSensor.NewServer(client, sensors.Devices))
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
	log.Println("Sensor server stopped")
}
