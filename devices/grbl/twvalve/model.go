package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jt05610/petri/grbl"
)

type TwoPositionThreeWayValve struct {
	txCh   chan []byte
	rxCh   <-chan []byte
	status *grbl.Status
}

func (d *TwoPositionThreeWayValve) UpdateStatus(status grbl.StatusUpdate) {
	if upd, ok := status.(*grbl.Status); ok {
		d.status = upd
		return
	}
}

func (d *TwoPositionThreeWayValve) Listen(ctx context.Context) error {
	buf := new(bytes.Buffer)
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-d.rxCh:
			buf.Write(msg)
			if msg[len(msg)-1] == '\n' {
				parser := grbl.NewParser(buf)
				upd, err := parser.Parse()
				if err != nil {
					fmt.Print(string(msg))
					panic(err)
				}
				d.UpdateStatus(upd)
				buf.Reset()
			}
		}
	}
}

type OpenARequest struct {
}

type OpenAResponse struct {
}

type OpenBRequest struct {
}

type OpenBResponse struct {
}
