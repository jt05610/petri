package sensor

import (
	"bytes"
	"context"
	"core/sensor/v1/sensor"
	"encoding/binary"
	"errors"
	"gonum.org/v1/gonum/stat"
	"modbus/v1/modbus"
	"sync"
	"time"
)

type regMap struct {
	InputRegisters   map[string]*modbus.InputRegister
	HoldingRegisters map[string]*modbus.HoldingRegister
	Coils            map[string]*modbus.Coil
	DiscreteInputs   map[string]*modbus.DiscreteInput
}

type device struct {
	sensor      *ModbusSensor
	RegisterMap *regMap
	tareValue   float32
}

type server struct {
	modbus.ModbusClient
	sensors map[string]*device
	mu      sync.Mutex
	sensor.UnimplementedDeviceServer
}

func (s *server) Tare(ctx context.Context, request *sensor.TareRequest) (*sensor.TareResponse, error) {
	dev := s.sensors[request.Id]
	if dev == nil {
		return nil, errors.New("device not found")
	}
	total := float32(0)
	var samples int32 = 10
	if request.Samples == nil {
		samples = *request.Samples
	}
	values := make([]float64, samples)
	readRequest := &sensor.ReadRequest{Id: request.Id}
	for i := int32(0); i < samples; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(100) * time.Millisecond):
			f, err := s.Read(ctx, readRequest)
			if err != nil {
				panic(err)
			}
			total += f.Value
			values[i] = float64(f.Value)
		}
	}
	dev.tareValue = total / float32(samples)
	return &sensor.TareResponse{
		Variance: float32(stat.Variance(values, nil)),
	}, nil
}

func (s *server) Read(ctx context.Context, request *sensor.ReadRequest) (*sensor.ReadResponse, error) {
	s.mu.Lock()
	dev := s.sensors[request.Id]
	s.mu.Unlock()
	if dev == nil {
		return nil, errors.New("device not found")
	}
	req := &modbus.ReadInputRegistersRequest{
		UnitId:       uint32(*dev.sensor.Sensor.UnitId),
		StartAddress: dev.RegisterMap.InputRegisters[dev.sensor.Kind].Address,
		Quantity:     dev.RegisterMap.InputRegisters[dev.sensor.Kind].Size,
	}
	res, err := s.ModbusClient.ReadInputRegisters(ctx, req)
	if err != nil {
		return nil, err
	}
	switch req.Quantity {
	case 1:
		var val uint16
		err := binary.Read(bytes.NewReader(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		cal := dev.sensor.Sensor.Calibration
		calVal := (float64(val) - cal.Intercept) / cal.Slope
		return &sensor.ReadResponse{
			Value: float32(calVal) - dev.tareValue,
		}, nil

	case 2:
		var val uint32
		err := binary.Read(bytes.NewReader(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		cal := dev.sensor.Sensor.Calibration
		calVal := (float64(val) - cal.Intercept) / cal.Slope
		return &sensor.ReadResponse{
			Value: float32(calVal) - dev.tareValue,
		}, nil
	default:
		return nil, errors.New("unsupported quantity")
	}
}

func NewServer(client modbus.ModbusClient, sensors []*ModbusSensor) sensor.DeviceServer {
	srv := &server{
		ModbusClient: client,
		sensors:      make(map[string]*device),
	}
	for _, s := range sensors {
		srv.sensors[s.Sensor.Id] = &device{
			sensor:      s,
			RegisterMap: &regMap{},
		}
		if s.RegisterMap.InputRegisters != nil {
			srv.sensors[s.Sensor.Id].RegisterMap.InputRegisters = make(map[string]*modbus.InputRegister)
			for _, r := range s.RegisterMap.InputRegisters {
				srv.sensors[s.Sensor.Id].RegisterMap.InputRegisters[r.Name] = r
			}
		}
		if s.RegisterMap.HoldingRegisters != nil {
			srv.sensors[s.Sensor.Id].RegisterMap.HoldingRegisters = make(map[string]*modbus.HoldingRegister)
			for _, r := range s.RegisterMap.HoldingRegisters {
				srv.sensors[s.Sensor.Id].RegisterMap.HoldingRegisters[r.Name] = r
			}
		}
		if s.RegisterMap.Coils != nil {
			srv.sensors[s.Sensor.Id].RegisterMap.Coils = make(map[string]*modbus.Coil)
			for _, r := range s.RegisterMap.Coils {
				srv.sensors[s.Sensor.Id].RegisterMap.Coils[r.Name] = r
			}
		}
		if s.RegisterMap.DiscreteInputs != nil {
			srv.sensors[s.Sensor.Id].RegisterMap.DiscreteInputs = make(map[string]*modbus.DiscreteInput)
			for _, r := range s.RegisterMap.DiscreteInputs {
				srv.sensors[s.Sensor.Id].RegisterMap.DiscreteInputs[r.Name] = r
			}
		}
	}
	return srv
}
