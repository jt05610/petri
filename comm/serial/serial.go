package serial

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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
	err = p.SetReadTimeout(time.Duration(1000) * time.Millisecond)
	if err != nil {
		return nil, err
	}
	err = p.ResetOutputBuffer()
	if err != nil {
		return nil, err
	}
	ret := &Port{port: p}
	return ret, nil
}

func (p *Port) Flush() error {
	buf := make([]byte, 1024)
	for {
		n, err := p.port.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}
	return nil
}
func (p *Port) Close() error {
	return p.port.Close()
}

func (p *Port) WritePort(data []byte) (int, error) {
	return p.port.Write(data)
}

func (p *Port) ReadPort(data []byte) (int, error) {
	return p.port.Read(data)
}

func equalToOneOf(b byte, bb []byte) bool {
	for _, v := range bb {
		if b == v {
			return true
		}
	}
	return false
}
func MultiSplit(split []byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if equalToOneOf(data[i], split) {
				return i + 1, data[:i], nil
			}
		}
		if !atEOF {
			return 0, nil, nil
		}
		return 0, data, bufio.ErrFinalToken
	}
}

func (p *Port) ChannelPort(ctx context.Context, writeCh <-chan []byte, split ...byte) (<-chan io.Reader, error) {
	p.rxCh = make(chan io.Reader, 100) // Buffer size can be adjusted as per your requirement
	go func() {
		scanner := bufio.NewScanner(p.port)
		if len(split) > 0 {
			scanner.Split(MultiSplit(split))
		}
		for {
			select {
			case <-ctx.Done():
				fmt.Println("closing listen port ")
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
