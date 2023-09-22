package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

func startMsg(props []uint16) []byte {
	if len(props) != 8 {
		panic("Invalid number of props")
	}
	ret := bytes.NewBuffer([]byte("R"))
	for _, prop := range props {
		err := binary.Write(ret, binary.LittleEndian, prop)
		if err != nil {
			panic(err)
		}
	}
	ret.WriteString("\n")
	return ret.Bytes()

}

func setPeriodMsg(period uint16) []byte {
	ret := bytes.NewBuffer([]byte("P"))
	err := binary.Write(ret, binary.LittleEndian, period)
	if err != nil {
		panic(err)
	}
	ret.WriteString("\n")
	return ret.Bytes()
}

func (d *MixingValve) Start(ctx context.Context, props []uint16) error {
	msg := startMsg(props)
	d.txCh <- msg
	return nil
}

func fromString(s string) []uint16 {
	commaSplit := strings.Split(s, ",")
	if len(commaSplit) != 8 {
		msg := fmt.Errorf("Invalid number of props: %d\nsent: %s", len(commaSplit), s)
		panic(msg)
	}
	ret := make([]uint16, 8)
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
	err := d.Start(ctx, props)
	if err != nil {
		return nil, err
	}
	return &MixResponse{}, nil
}

func (d *MixingValve) Stop(ctx context.Context) error {
	msg := []byte("S\n")
	d.txCh <- msg
	return nil
}

func (d *MixingValve) Mixed(ctx context.Context, req *MixedRequest) (*MixedResponse, error) {
	err := d.Stop(ctx)
	if err != nil {
		return nil, err
	}
	return &MixedResponse{}, nil
}
