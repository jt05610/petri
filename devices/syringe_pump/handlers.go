package main

import (
	"context"
	"fmt"
	modbus "github.com/jt05610/petri/proto/v1"
	"log"
	"time"
)

type Coil uint32

const (
	Start Coil = iota
	Stop
	Home
	MoveToStall
	SetZero
	Enable
)

func (d *SyringePump) writeCoilRequest(ctx context.Context, coil Coil, value bool) error {
	_, err := d.client.WriteSingleCoil(ctx, &modbus.WriteSingleCoilRequest{
		UnitId:  d.unitID,
		Address: uint32(coil),
		Value:   value,
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *SyringePump) start(ctx context.Context) error {
	return d.writeCoilRequest(ctx, Start, true)
}

func (d *SyringePump) stop(ctx context.Context) error {
	return d.writeCoilRequest(ctx, Stop, true)
}

func (d *SyringePump) home(ctx context.Context) error {
	return d.writeCoilRequest(ctx, Home, true)
}

func (d *SyringePump) moveToStall(ctx context.Context) error {
	return d.writeCoilRequest(ctx, MoveToStall, true)
}

func (d *SyringePump) setZero(ctx context.Context) error {
	return d.writeCoilRequest(ctx, SetZero, true)
}

func (d *SyringePump) enable(ctx context.Context) error {
	return d.writeCoilRequest(ctx, Enable, true)
}

func (d *SyringePump) disable(ctx context.Context) error {
	return d.writeCoilRequest(ctx, Enable, false)
}

func (d *SyringePump) writeRegisterRequest(ctx context.Context, register HoldingRegister, value uint32) error {
	_, err := d.client.WriteSingleRegister(ctx, &modbus.WriteSingleRegisterRequest{
		UnitId:  d.unitID,
		Address: uint32(register),
		Value:   value,
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *SyringePump) setTargetPosition(ctx context.Context, value uint32) error {
	return d.writeRegisterRequest(ctx, TargetPosition, value)
}

func (d *SyringePump) setTargetVelocity(ctx context.Context, value uint32) error {
	return d.writeRegisterRequest(ctx, TargetVelocity, value)
}

func (d *SyringePump) setMoveTo(ctx context.Context, value uint32) error {
	return d.writeRegisterRequest(ctx, MoveTo, value)
}

func (d *SyringePump) setAccel(ctx context.Context, value uint32) error {
	return d.writeRegisterRequest(ctx, Accel, value)
}

func (d *SyringePump) setStallGuard(ctx context.Context, value uint32) error {
	return d.writeRegisterRequest(ctx, StallGuard, value)
}

func getRegisterValue(bb []byte) (uint32, error) {
	// bb[0] tells us how many runes make up the value in the register
	switch bb[0] {
	case 1:
		return uint32(bb[1])<<8 + uint32(bb[2]), nil
	case 2:
		return uint32(bb[1])<<24 + uint32(bb[2])<<16 + uint32(bb[3])<<8 + uint32(bb[4]), nil
	default:
		return 0, fmt.Errorf("invalid number of runes in register")
	}
}

func (d *SyringePump) readRegisterRequest(ctx context.Context, register HoldingRegister) (uint32, error) {
	res, err := d.client.ReadHoldingRegisters(ctx, &modbus.ReadHoldingRegistersRequest{
		UnitId:       d.unitID,
		StartAddress: uint32(register),
		Quantity:     1,
	})
	if err != nil {
		return 0, err
	}
	return getRegisterValue(res.Data)
}

func (d *SyringePump) readInputRegisterRequest(ctx context.Context, register InputRegister) (uint32, error) {
	res, err := d.client.ReadInputRegisters(ctx, &modbus.ReadInputRegistersRequest{
		UnitId:       d.unitID,
		StartAddress: uint32(register),
		Quantity:     1,
	})
	if err != nil {
		return 0, err
	}
	return getRegisterValue(res.Data)
}

func (d *SyringePump) getTargetPosition(ctx context.Context) (uint32, error) {
	return d.readRegisterRequest(ctx, TargetPosition)
}

func (d *SyringePump) getTargetVelocity(ctx context.Context) (uint32, error) {
	return d.readRegisterRequest(ctx, TargetVelocity)
}

func (d *SyringePump) getMoveTo(ctx context.Context) (uint32, error) {
	return d.readRegisterRequest(ctx, MoveTo)
}

func (d *SyringePump) getAccel(ctx context.Context) (uint32, error) {
	return d.readRegisterRequest(ctx, Accel)
}

func (d *SyringePump) getStallGuard(ctx context.Context) (uint32, error) {
	return d.readRegisterRequest(ctx, StallGuard)
}

func (d *SyringePump) getCurrentPosition(ctx context.Context) (uint32, error) {
	return d.readInputRegisterRequest(ctx, CurrentPosition)
}

func (d *SyringePump) getCurrentVelocity(ctx context.Context) (uint32, error) {
	return d.readInputRegisterRequest(ctx, CurrentVelocity)
}

func (d *SyringePump) getTStep(ctx context.Context) (uint32, error) {
	return d.readInputRegisterRequest(ctx, TStep)
}

func (d *SyringePump) getForce(ctx context.Context) (uint32, error) {
	return d.readInputRegisterRequest(ctx, Force)
}

func (d *SyringePump) readCoilRequest(ctx context.Context, coil Coil) (bool, error) {
	res, err := d.client.ReadCoils(ctx, &modbus.ReadCoilsRequest{
		UnitId:       d.unitID,
		StartAddress: uint32(coil),
		Quantity:     1,
	})
	if err != nil {
		return false, err
	}
	return res.Data[0], nil
}

func (d *SyringePump) getEnable(ctx context.Context) (bool, error) {
	return d.readCoilRequest(ctx, Enable)
}

func (d *SyringePump) readDiscreteInputRequest(ctx context.Context, input DiscreteInput) (bool, error) {
	res, err := d.client.ReadDiscreteInputs(ctx, &modbus.ReadDiscreteInputsRequest{
		UnitId:       d.unitID,
		StartAddress: uint32(input),
		Quantity:     1,
	})
	if err != nil {
		return false, err
	}
	return res.Data[0], nil
}

func (d *SyringePump) getIsMoving(ctx context.Context) (bool, error) {
	return d.readDiscreteInputRequest(ctx, IsMoving)
}

type DiscreteInput uint32

const (
	IsMoving DiscreteInput = iota
)

type HoldingRegister uint32

const (
	TargetPosition HoldingRegister = iota
	TargetVelocity
	MoveTo
	Accel
	StallGuard
)

type InputRegister uint32

const (
	CurrentPosition InputRegister = iota
	CurrentVelocity
	TStep
	Force
)

func (d *SyringePump) startVelChannel(ctx context.Context) error {
	d.velCh = make(chan float64)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(d.increment):
				v, err := d.getCurrentVelocity(ctx)
				if err != nil {
					log.Printf("error getting current velocity: %v", err)
					continue
				}
				d.velCh <- float64(v)
			}
		}
	}()
	return nil
}

func (d *SyringePump) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	// Set syringe diameter
	// Set syringe volume
	// Set pump calibration
	// Set pump microstep
	// Set pump rate increment
	// Set unitID

	// home pump
	err := d.home(ctx)
	if err != nil {
		return nil, err
	}
	// set max pos
	// find max pos
	return &InitializeResponse{}, nil
}

func (d *SyringePump) Pump(ctx context.Context, req *PumpRequest) (*PumpResponse, error) {
	return &PumpResponse{}, nil
}

func (d *SyringePump) Stop(ctx context.Context, req *StopRequest) (*StopResponse, error) {
	return &StopResponse{}, nil
}
