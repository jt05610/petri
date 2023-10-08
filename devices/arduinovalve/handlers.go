package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
)

func startMsg(props []uint16) []byte {
	if len(props) != 7 {
		panic("Invalid number of props")
	}
	ret := bytes.NewBuffer([]byte("R"))
	for i, prop := range props {
		propStr := strconv.Itoa(int(prop))
		ret.WriteString(propStr)
		if i != len(props)-1 {
			ret.WriteString(",")
		}
	}
	ret.WriteString("\n")
	bb := ret.Bytes()
	fmt.Printf("startMsg: %s\n", strconv.Quote(string(bb)))
	return bb
}

func setPeriodMsg(period uint16) []byte {
	ret := bytes.NewBuffer([]byte("P"))
	periodStr := strconv.Itoa(int(period))
	ret.WriteString(periodStr)
	ret.WriteString("\n")
	bb := ret.Bytes()
	fmt.Printf("setPeriodMsg: %s\n", strconv.Quote(string(bb)))
	return bb
}

func (d *MixingValve) do(ctx context.Context, msg []byte) error {
	fmt.Printf("sending: %s\n", strconv.Quote(string(msg)))
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if d.cts.Load() {
				d.cts.Store(false)
				d.txCh <- msg
				fmt.Printf("sent: %s\n", strconv.Quote(string(msg)))
				return nil
			}
		}
	}
}

func (d *MixingValve) Start(ctx context.Context, props []uint16) error {
	msg := startMsg(props)
	return d.do(ctx, msg)
}

func (d *MixingValve) SetPeriod(ctx context.Context, period uint16) error {
	msg := setPeriodMsg(period)
	return d.do(ctx, msg)
}

func fromString(s string) []uint16 {
	commaSplit := strings.Split(s, ",")
	if len(commaSplit) != 7 {
		msg := fmt.Errorf("Invalid number of props: %d\nsent: %s", len(commaSplit), s)
		panic(msg)
	}
	ret := make([]uint16, 7)
	for i, prop := range commaSplit {
		propInt, err := strconv.Atoi(prop)
		if err != nil {
			panic(err)
		}
		if propInt > (1<<16)-1 {
			panic("Invalid prop")
		}
		ret[i] = uint16(propInt)
	}
	return ret
}

func (d *MixingValve) Mix(ctx context.Context, req *MixRequest) (*MixResponse, error) {
	props := fromString(req.Proportions)
	period := uint16(req.Period)
	currentPeriod := d.period.Load()
	if uint16(currentPeriod) != period {
		d.period.Store(int32(period))
		err := d.SetPeriod(ctx, period)
		if err != nil {
			return nil, err
		}
	}

	err := d.Start(ctx, props)
	if err != nil {
		return nil, err
	}
	return &MixResponse{}, nil
}

func (d *MixingValve) Stop(ctx context.Context) error {
	msg := []byte("S\n")
	return d.do(ctx, msg)
}

func (d *MixingValve) Mixed(ctx context.Context, req *MixedRequest) (*MixedResponse, error) {
	return &MixedResponse{}, nil
}
