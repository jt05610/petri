package wire

import (
	"errors"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type SerialOpt struct {
	Port     string
	VID      string
	PID      string
	Baud     int
	Parity   serial.Parity
	DataBits int
	StopBits serial.StopBits
}

type Serial struct {
	rx serial.Port
	tx serial.Port
}

var DefaultSerial = &SerialOpt{
	Port:     "",
	VID:      "1A86",
	PID:      "7523",
	Baud:     19200,
	Parity:   serial.NoParity,
	DataBits: 8,
	StopBits: serial.TwoStopBits,
}

var NoPort = errors.New("no port given to NewSerial")

func NewSerial(opts ...*SerialOpt) (*Serial, error) {
	opt := DefaultSerial
	if opts != nil {
		opt = opts[0]
	}
	s := &Serial{}
	var err error
	if opt.Port == "" {
		if opt.PID == "" && opt.VID == "" {
			return nil, NoPort
		}
		ports, err := enumerator.GetDetailedPortsList()
		if err != nil {
			panic(err)
		}
		vCheck := len(opt.PID) > 0
		for _, port := range ports {
			if port.PID == opt.PID {
				opt.Port = port.Name
				if vCheck {
					if port.VID == opt.VID {
						opt.Port = port.Name
						break
					}
				} else {
					break
				}
			}
		}
	}
	if len(opt.Port) == 0 {
		panic(errors.New("no port found"))
	}
	mode := &serial.Mode{
		BaudRate: opt.Baud,
		Parity:   opt.Parity,
		DataBits: opt.DataBits,
		StopBits: opt.StopBits,
	}
	var portName string
	if opt.Port[:8] == "/dev/cu." {
		portName = opt.Port[8:]
	} else {
		portName = opt.Port[9:]
	}

	s.rx, err = serial.Open("/dev/cu."+portName, mode)

	if err != nil {
		panic(err)
	}
	s.tx = s.rx
	return s, err
}

func (s *Serial) Read(p []byte) (n int, err error) {
	return s.rx.Read(p)
}

func (s *Serial) Write(p []byte) (n int, err error) {
	return s.tx.Write(p)
}
