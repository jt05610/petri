package main

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"modbus/metrics"
	"modbus/pdu"
	"modbus/serial"
	axCmd "modbus/v1/axis/cmd"
	"modbus/v1/modbus"
	sensorCmd "modbus/v1/sensor/cmd"
	"modbus/wire"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	metricsAddr = flag.String("metrics", "localhost:55155", "metrics listen address")
	modAddr     = flag.String("server", "localhost:55055", "rpc server listen address")
	axisAddr    = flag.String("axis", "localhost:55056", "axis server listen address")
	sensorAddr  = flag.String("sensor", "localhost:55057", "sensor server listen address")
	certFile    = flag.String("cert", "cert.pem", "tls cert file")
	keyFile     = flag.String("key", "key.pem", "tls key file")
)

type modbusServer struct {
	modbus.UnimplementedModbusServer
	client *serial.Client
}

func (s *modbusServer) ReadCoils(ctx context.Context, input *modbus.ReadCoilsRequest) (*modbus.ReadCoilsResponse, error) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	req := pdu.ReadCoils(uint16(input.StartAddress), uint16(input.Quantity))
	resp, err := s.client.Request(ctx, uint8(input.UnitId), req)
	metrics.SerialRequests.Add(1)
	if err != nil {
		metrics.NoResponseErrors.Add(1)
		return nil, err
	}
	ret := &modbus.ReadCoilsResponse{
		Data: make([]bool, 0, len(resp.Data)),
	}
	for _, b := range resp.Data {
		ret.Data = append(ret.Data, b == 0xFF)
	}
	return ret, nil
}

func (s *modbusServer) ReadDiscreteInputs(ctx context.Context, input *modbus.ReadDiscreteInputsRequest) (*modbus.ReadDiscreteInputsResponse, error) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	req := pdu.ReadDiscreteInputs(uint16(input.StartAddress), uint16(input.Quantity))
	resp, err := s.client.Request(ctx, uint8(input.UnitId), req)
	metrics.SerialRequests.Add(1)
	if err != nil {
		metrics.NoResponseErrors.Add(1)
		return nil, err
	}
	ret := &modbus.ReadDiscreteInputsResponse{
		Data: make([]bool, 0, len(resp.Data)),
	}
	for _, b := range resp.Data {
		ret.Data = append(ret.Data, b == 0xFF)
	}
	return ret, nil
}

func (s *modbusServer) ReadHoldingRegisters(ctx context.Context, input *modbus.ReadHoldingRegistersRequest) (*modbus.ReadHoldingRegistersResponse, error) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	req := pdu.ReadHoldingRegisters(uint16(input.StartAddress), uint16(input.Quantity))
	resp, err := s.client.Request(ctx, uint8(input.UnitId), req)
	metrics.SerialRequests.Add(1)
	if err != nil {
		metrics.NoResponseErrors.Add(1)
		return nil, err
	}
	ret := &modbus.ReadHoldingRegistersResponse{
		Data: resp.Data,
	}

	return ret, nil
}

func (s *modbusServer) ReadInputRegisters(ctx context.Context, input *modbus.ReadInputRegistersRequest) (*modbus.ReadInputRegistersResponse, error) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	req := pdu.ReadInputRegisters(uint16(input.StartAddress), uint16(input.Quantity))
	resp, err := s.client.Request(ctx, uint8(input.UnitId), req)
	metrics.SerialRequests.Add(1)
	if err != nil {
		metrics.NoResponseErrors.Add(1)
		return nil, err
	}
	ret := &modbus.ReadInputRegistersResponse{
		Data: resp.Data,
	}
	return ret, nil
}

func (s *modbusServer) WriteSingleCoil(ctx context.Context, input *modbus.WriteSingleCoilRequest) (*modbus.WriteSingleCoilResponse, error) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	var val uint16
	if input.Value {
		val = 0xFF00
	}
	req := pdu.WriteCoil(uint16(input.Address), val)
	_, err := s.client.Request(ctx, uint8(input.UnitId), req)
	metrics.SerialRequests.Add(1)
	if err != nil {
		metrics.NoResponseErrors.Add(1)
		return nil, err
	}
	return &modbus.WriteSingleCoilResponse{
		Address: input.Address,
		Value:   input.Value,
	}, nil
}

func (s *modbusServer) WriteSingleRegister(ctx context.Context, input *modbus.WriteSingleRegisterRequest) (*modbus.WriteSingleRegisterResponse, error) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())
	req := pdu.WriteRegister(uint16(input.Address), uint16(input.Value))
	_, err := s.client.Request(ctx, uint8(input.UnitId), req)
	metrics.SerialRequests.Add(1)
	if err != nil {
		metrics.NoResponseErrors.Add(1)
		return nil, err
	}
	return &modbus.WriteSingleRegisterResponse{
		Address: input.Address,
		Value:   input.Value,
	}, nil
}

func newServer() *modbusServer {
	// Start mock modbus client
	ser, err := wire.NewSerial()
	if err != nil {
		panic(err)
	}
	dl := serial.NewDataLink(ser)
	modClient := &serial.Client{DataLink: dl}
	return &modbusServer{client: modClient}
}

func newHTTPServer(ctx context.Context, cfg *Config, mux http.Handler,
	stateFunc func(net.Conn, http.ConnState)) error {
	l, err := net.Listen("tcp", cfg.MetricsAddr)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              cfg.MetricsAddr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
		ConnState:         stateFunc,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
			log.Printf("🚌📊 server shutdown error: %v\n", err)
		}
		log.Printf("🚌📊 server shutdown\n")
	}()
	log.Printf("🚌📊 server listening on https://%s/metrics\n", cfg.MetricsAddr)
	return srv.ServeTLS(l, cfg.CertFile, cfg.KeyFile)
}

func connStateMetrics(_ net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		metrics.OpenConnections.Add(1)
	case http.StateClosed:
		metrics.OpenConnections.Add(-1)
	}
}

func runModbusServer(ctx context.Context, cfg *Config) {
	var wg sync.WaitGroup
	// Launching metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Launching rpc server
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		panic(err)
	}
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			log.Printf("🚌📊 server shutdown\n")
		}()
		if err := newHTTPServer(ctx, cfg, mux, connStateMetrics); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", cfg.ModbusAddr)
	if err != nil {
		panic(err)
	}
	log.Printf("🚌 server listening on https://%s/", cfg.ModbusAddr)
	var opts []grpc.ServerOption
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer(opts...)
	modbus.RegisterModbusServer(grpcServer, newServer())
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()
	if err := grpcServer.Serve(
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
	log.Println("🚌 server shutdown")
	wg.Wait()
}

type Config struct {
	MetricsAddr string
	AxisAddr    string
	ModbusAddr  string
	SensorAddr  string
	GraphQLAddr string
	CertFile    string
	KeyFile     string
}

func runAxisServer(ctx context.Context, cfg *Config) {
	axCmd.Execute(ctx, cfg.AxisAddr, cfg.ModbusAddr, cfg.CertFile, cfg.KeyFile)
}

func runSensorServer(ctx context.Context, cfg *Config) {
	sensorCmd.Execute(ctx, cfg.SensorAddr, cfg.ModbusAddr, cfg.CertFile, cfg.KeyFile)
}

func main() {
	flag.Parse()
	cfg := &Config{
		MetricsAddr: *metricsAddr,
		AxisAddr:    *axisAddr,
		ModbusAddr:  *modAddr,
		SensorAddr:  *sensorAddr,
		CertFile:    *certFile,
		KeyFile:     *keyFile,
	}
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		runAxisServer(ctx, cfg)
	}()
	go func() {
		defer wg.Done()
		runSensorServer(ctx, cfg)
	}()
	go func() {
		defer wg.Done()
		runModbusServer(ctx, cfg)
	}()
	sigint := make(chan os.Signal, 1)
	defer close(sigint)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	cancel()
	wg.Wait()
}
