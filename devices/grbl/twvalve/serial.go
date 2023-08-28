package main

import (
	"context"
	"go.bug.st/serial"
	"strings"
	"sync"
	"time"
)

type Port struct {
	port serial.Port
	mu   sync.RWMutex
	rxCh chan []byte
	txCh chan []byte
}

func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	return ports, nil
}

func (p *Port) Do(ctx context.Context, lines string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	splitLines := strings.Split(lines, "\n")
	for _, line := range splitLines {
		p.txCh <- []byte(line + "\n")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-p.rxCh:
			if strings.Contains(string(msg), "ok") {
				continue
			}
		}
	}
	return nil
}

func OpenPort(port string, baud int) (*Port, error) {
	p, err := serial.Open(port, &serial.Mode{
		BaudRate: baud,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}

	err = p.SetReadTimeout(time.Duration(500) * time.Millisecond)
	if err != nil {
		return nil, err
	}
	return &Port{port: p}, nil
}

func (p *Port) Close() error {
	return p.port.Close()
}

func WritePort(port serial.Port, data []byte) (int, error) {
	return port.Write(data)
}

func ReadPort(port serial.Port, data []byte) (int, error) {
	return port.Read(data)
}

func (p *Port) ChannelPort(ctx context.Context, writeCh <-chan []byte) (<-chan []byte, error) {
	p.rxCh = make(chan []byte, 1024) // Buffer size can be adjusted as per your requirement
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			buf := make([]byte, 128)
			n, err := p.port.Read(buf)
			if err != nil {
				panic(err)
			}
			if n > 0 {
				p.rxCh <- buf[:n]
			} else {
				continue
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-writeCh:
				if _, err := p.port.Write(data); err != nil {
					// handle the error
					return
				}
			}
		}
	}()

	return p.rxCh, nil
}
