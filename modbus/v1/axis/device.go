package axis

import (
	"bytes"
	"context"
	"core/axis/v1/axis"
	"encoding/binary"
	"errors"
	"modbus/v1/modbus"
	"strings"
	"sync"
)

type regMap struct {
	InputRegisters   map[string]*modbus.InputRegister
	HoldingRegisters map[string]*modbus.HoldingRegister
	Coils            map[string]*modbus.Coil
	DiscreteInputs   map[string]*modbus.DiscreteInput
}
type device struct {
	ax          *ModbusAxis
	RegisterMap *regMap
}
type service struct {
	modbus.ModbusClient
	devices map[string]*device
	mu      sync.Mutex
	axis.UnimplementedDeviceServer
}

func (s *service) getAxis(id string) (*device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.devices[id]
	if !ok {
		return nil, errors.New("device not found")
	}

	return d, nil
}

func irRequest(id uint32, r *modbus.InputRegister) *modbus.ReadInputRegistersRequest {
	return &modbus.ReadInputRegistersRequest{
		UnitId:       id,
		StartAddress: r.Address,
		Quantity:     r.Size,
	}
}

func hrReadRequest(id uint32, r *modbus.HoldingRegister) *modbus.ReadHoldingRegistersRequest {
	return &modbus.ReadHoldingRegistersRequest{
		UnitId:       id,
		StartAddress: r.Address,
		Quantity:     r.Size,
	}
}

func hrWriteRequest(id uint32, r *modbus.HoldingRegister, val uint16) *modbus.WriteSingleRegisterRequest {
	return &modbus.WriteSingleRegisterRequest{
		UnitId:  id,
		Address: r.Address,
		Value:   uint32(val),
	}
}

func coilWriteRequest(id uint32, r *modbus.Coil, val bool) *modbus.WriteSingleCoilRequest {
	return &modbus.WriteSingleCoilRequest{
		UnitId:  id,
		Address: r.Address,
		Value:   val,
	}
}

func (s *service) GetPosition(ctx context.Context, req *axis.GetPositionRequest) (*axis.GetPositionResponse, error) {
	d, err := s.getAxis(req.Id)
	if err != nil {
		return nil, err
	}
	mbReq := irRequest(uint32(*d.ax.Axis.UnitId), d.RegisterMap.InputRegisters["current_pos"])
	res, err := s.ReadInputRegisters(ctx, mbReq)
	if err != nil {
		return nil, err
	}
	switch mbReq.Quantity {
	case 1:
		var val uint16
		err = binary.Read(bytes.NewBuffer(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		return &axis.GetPositionResponse{
			Position: &axis.Position{
				Value: float64(val),
			},
		}, nil
	case 2:
		var val uint32
		err = binary.Read(bytes.NewBuffer(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		return &axis.GetPositionResponse{
			Position: &axis.Position{
				Value: float64(val),
			},
		}, nil
	default:
		return nil, errors.New("unsupported quantity")
	}
}

func (s *service) GetVelocity(ctx context.Context, req *axis.GetVelocityRequest) (*axis.GetVelocityResponse, error) {
	d, err := s.getAxis(req.Id)
	if err != nil {
		return nil, err
	}
	mbReq := irRequest(uint32(*d.ax.Axis.UnitId), d.RegisterMap.InputRegisters["current_vel"])
	res, err := s.ReadInputRegisters(ctx, mbReq)
	switch mbReq.Quantity {
	case 1:
		var val uint16
		err = binary.Read(bytes.NewBuffer(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		return &axis.GetVelocityResponse{
			Velocity: &axis.Velocity{
				Value: float64(val),
			},
		}, nil
	case 2:
		var val uint32
		err = binary.Read(bytes.NewBuffer(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		return &axis.GetVelocityResponse{
			Velocity: &axis.Velocity{
				Value: float64(val),
			},
		}, nil
	default:
		return nil, errors.New("unsupported quantity")
	}
}

func (s *service) GetAcceleration(ctx context.Context, req *axis.GetAccelerationRequest) (*axis.GetAccelerationResponse, error) {
	d, err := s.getAxis(req.Id)
	if err != nil {
		return nil, err
	}
	mbReq := hrReadRequest(uint32(*d.ax.Axis.UnitId), d.RegisterMap.HoldingRegisters["accel"])

	res, err := s.ReadHoldingRegisters(ctx, mbReq)
	if err != nil {
		return nil, err
	}
	switch mbReq.Quantity {
	case 1:
		var val uint16
		err = binary.Read(bytes.NewBuffer(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		return &axis.GetAccelerationResponse{
			Acceleration: &axis.Acceleration{
				Value: float64(val),
			},
		}, nil
	case 2:
		var val uint32
		err = binary.Read(bytes.NewBuffer(res.Data), binary.BigEndian, &val)
		if err != nil {
			return nil, err
		}
		return &axis.GetAccelerationResponse{
			Acceleration: &axis.Acceleration{
				Value: float64(val),
			},
		}, nil
	default:
		return nil, errors.New("unsupported quantity")
	}
}

func (s *service) MoveTo(ctx context.Context, req *axis.MoveToRequest) (*axis.MoveToResponse, error) {
	d, err := s.getAxis(req.Id)
	if err != nil {
		return nil, err
	}
	reqs := make([]*modbus.WriteSingleRegisterRequest, 0, 3)
	if req.Position != nil {
		reqs = append(reqs, hrWriteRequest(uint32(*d.ax.Axis.UnitId), d.RegisterMap.HoldingRegisters["target_pos"], uint16(req.Position.Value)))

	}
	if req.Velocity != nil {
		reqs = append(reqs, hrWriteRequest(uint32(*d.ax.Axis.UnitId), d.RegisterMap.HoldingRegisters["target_vel"], uint16(req.Velocity.Value)))
	}
	if req.Acceleration != nil {
		reqs = append(reqs, hrWriteRequest(uint32(*d.ax.Axis.UnitId), d.RegisterMap.HoldingRegisters["target_accel"], uint16(req.Acceleration.Value)))
	}
	for _, req := range reqs {
		_, err := s.WriteSingleRegister(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	_, err = s.Start(ctx, &axis.StartRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &axis.MoveToResponse{
		Id:           req.Id,
		Position:     req.Position,
		Velocity:     req.Velocity,
		Acceleration: req.Acceleration,
	}, nil
}

func (s *service) SetPosition(ctx context.Context, request *axis.SetPositionRequest) (*axis.SetPositionResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := hrWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.HoldingRegisters["target_pos"], uint16(request.Position.Value))
	_, err = s.WriteSingleRegister(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.SetPositionResponse{
		Id: request.Id,
	}, nil
}

func (s *service) MoveToStall(ctx context.Context, request *axis.MoveToStallRequest) (*axis.MoveToStallResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := coilWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.Coils["movetostall"], true)
	_, err = s.WriteSingleCoil(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.MoveToStallResponse{
		Id: request.Id,
	}, nil
}

func (s *service) Start(ctx context.Context, request *axis.StartRequest) (*axis.StartResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := coilWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.Coils["start"], true)
	_, err = s.WriteSingleCoil(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.StartResponse{
		Id: request.Id,
	}, nil
}

func (s *service) Stop(ctx context.Context, request *axis.StopRequest) (*axis.StopResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := coilWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.Coils["stop"], true)
	_, err = s.WriteSingleCoil(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.StopResponse{
		Id: request.Id,
	}, nil
}

func (s *service) Home(ctx context.Context, request *axis.HomeRequest) (*axis.HomeResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := coilWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.Coils["home"], true)
	_, err = s.WriteSingleCoil(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.HomeResponse{
		Id: request.Id,
	}, nil
}

func (s *service) Enable(ctx context.Context, request *axis.EnableRequest) (*axis.EnableResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := coilWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.Coils["enable"], true)
	_, err = s.WriteSingleCoil(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.EnableResponse{
		Id: request.Id,
	}, nil
}

func (s *service) Disable(ctx context.Context, request *axis.DisableRequest) (*axis.DisableResponse, error) {
	dev, err := s.getAxis(request.Id)
	if err != nil {
		return nil, err
	}
	req := coilWriteRequest(uint32(*dev.ax.Axis.UnitId), dev.RegisterMap.Coils["enable"], false)
	_, err = s.WriteSingleCoil(ctx, req)
	if err != nil {
		return nil, err
	}
	return &axis.DisableResponse{
		Id: request.Id,
	}, nil
}

func NewServer(client modbus.ModbusClient, axes []*ModbusAxis) axis.DeviceServer {
	srv := &service{
		ModbusClient: client,
		devices:      make(map[string]*device, len(axes)),
	}
	for _, axs := range axes {
		if axs.Axis.UnitId == nil {
			panic("UnitId must not be nil for modbus axis, abandoning")
		}
		newDevice := &device{
			ax:          axs,
			RegisterMap: &regMap{},
		}
		if axs.RegisterMap.Coils != nil {
			newDevice.RegisterMap.Coils = make(map[string]*modbus.Coil, len(axs.RegisterMap.Coils))
		}
		if axs.RegisterMap.HoldingRegisters != nil {
			newDevice.RegisterMap.HoldingRegisters = make(map[string]*modbus.HoldingRegister, len(axs.RegisterMap.HoldingRegisters))
		}
		if axs.RegisterMap.InputRegisters != nil {
			newDevice.RegisterMap.InputRegisters = make(map[string]*modbus.InputRegister, len(axs.RegisterMap.InputRegisters))
		}
		if axs.RegisterMap.DiscreteInputs != nil {
			newDevice.RegisterMap.DiscreteInputs = make(map[string]*modbus.DiscreteInput, len(axs.RegisterMap.DiscreteInputs))
		}
		for _, v := range axs.RegisterMap.Coils {
			newDevice.RegisterMap.Coils[strings.ToLower(v.Name)] = v
		}
		for _, v := range axs.RegisterMap.HoldingRegisters {
			newDevice.RegisterMap.HoldingRegisters[strings.ToLower(v.Name)] = v
		}
		for _, v := range axs.RegisterMap.InputRegisters {
			newDevice.RegisterMap.InputRegisters[strings.ToLower(v.Name)] = v
		}
		for _, v := range axs.RegisterMap.DiscreteInputs {
			newDevice.RegisterMap.DiscreteInputs[strings.ToLower(v.Name)] = v
		}
		srv.devices[axs.Axis.Id] = newDevice
	}
	return srv
}
