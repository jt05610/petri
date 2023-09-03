package serial

import (
	"bufio"
	"bytes"
	"context"
	"go.bug.st/serial"
	"io"
	"sync"
	"time"
)

type Port struct {
	port serial.Port
	mu   sync.RWMutex
	rxCh chan io.Reader
	txCh chan []byte
}

func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	return ports, nil
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

func (p *Port) ChannelPort(ctx context.Context, writeCh <-chan []byte) (<-chan io.Reader, error) {
	p.rxCh = make(chan io.Reader, 100) // Buffer size can be adjusted as per your requirement
	go func() {
		scanner := bufio.NewScanner(p.port)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if scanner.Scan() {
					p.rxCh <- bytes.NewBuffer(scanner.Bytes())
				}
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
