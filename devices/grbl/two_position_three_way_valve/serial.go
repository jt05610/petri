package main

import (
	"context"
	"go.bug.st/serial"
)

func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	return ports, nil
}

func OpenPort(port string, baud int) (serial.Port, error) {
	p, err := serial.Open(port, &serial.Mode{
		BaudRate: baud,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ClosePort(port serial.Port) error {
	return port.Close()
}

func WritePort(port serial.Port, data []byte) (int, error) {
	return port.Write(data)
}

func ReadPort(port serial.Port, data []byte) (int, error) {
	return port.Read(data)
}

func ChannelPort(ctx context.Context, port serial.Port, writeCh <-chan []byte) (<-chan []byte, error) {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-writeCh:
				_, err := port.Write(data)
				if err != nil {
					panic(err)
				}
			default:
				buf := make([]byte, 128)
				n, err := port.Read(buf)
				if err != nil {
					panic(err)
				}
				if n > 0 {
					ch <- buf[:n]
				}
			}
		}
	}()
	return ch, nil
}
