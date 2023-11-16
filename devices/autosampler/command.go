package autosampler

import (
	"io"
	"strconv"
	"strings"
)

type Command interface {
	Bytes() []byte
}

type Service interface {
	Load(r io.Reader) (Response, error)
	Flush(w io.Writer, cmd Command) error
}

type Header byte

func (h Header) String() string {
	return string(h)
}

const (
	Get Header = 'G'
	Set Header = 'S'
)

type Request struct {
	Name string
	Header
	Id int
}

type Response interface {
	Bytes() []byte
}

func intString(i int) string {
	return strconv.Itoa(i)
}

func (r *Request) idString() string {
	return strconv.Itoa(r.Id)
}

func (r *Request) idBytes() []byte {
	return []byte(r.idString())
}

func (r *Request) Bytes() []byte {
	joined := strings.Join(
		[]string{
			r.Header.String(),
			r.idString(),
		},
		",",
	)
	return []byte(joined)
}

type (
	Connect struct{}
	Close   struct{}
	Wait    struct {
		Duration int
	}
	DeviceIDRequest        struct{}
	TrayStatusRequest      struct{}
	InjectionStatusRequest struct{}
	ExternalControl        struct{}
)

func (e *ExternalControl) Bytes() []byte {
	return []byte("E")
}

func (i *InjectionStatusRequest) Bytes() []byte {
	return []byte("G,13")
}

func (t *TrayStatusRequest) Bytes() []byte {
	return []byte("G,12")
}

func (d *DeviceIDRequest) Bytes() []byte {
	return []byte("G,10")
}

func (w *Wait) Bytes() []byte {
	return []byte(
		strings.Join(
			[]string{
				Set.String(),
				"33",
				intString(w.Duration),
			},
			",",
		),
	)
}

func (c *Close) Bytes() []byte {
	return []byte("3")
}

func (c *Connect) Bytes() []byte {
	return []byte("2")
}

var (
	_ Command = (*Request)(nil)
	_ Command = (*Connect)(nil)
	_ Command = (*Close)(nil)
	_ Command = (*Wait)(nil)
	_ Command = (*DeviceIDRequest)(nil)
	_ Command = (*TrayStatusRequest)(nil)
	_ Command = (*InjectionStatusRequest)(nil)
	_ Command = (*ExternalControl)(nil)
)

func InitializeCommands() []Command {
	return []Command{
		&DeviceIDRequest{},
		&Connect{},
		&Request{Name: "", Header: Set, Id: 29},
		&Wait{Duration: 2500},
		&Wait{Duration: 1000},
		&Wait{Duration: 500},
		&Wait{Duration: 250},
		&Wait{Duration: 100},
		&Wait{Duration: 50},
		&Request{Name: "", Header: Get, Id: 17},
	}
}
