package autosampler

import (
	"strconv"
	"strings"
)

type Header byte

func (h Header) String() string {
	return string(h)
}

const (
	None            Header = 'N'
	Get             Header = 'G'
	Set             Header = 'S'
	Inject          Header = 'I'
	ExternalControl Header = 'E'
)

type Request struct {
	Name string
	Header
	Id   int
	Data Byteable
}

type Byteable interface {
	Bytes() []byte
}

type StringData string

func (s StringData) Bytes() []byte {
	return []byte(s)
}

type IntData int

func (i IntData) Bytes() []byte {
	return []byte(intString(int(i)))
}

type Response struct {
	*Request
	Data Byteable
}

type IntArrayData []int

func (i IntArrayData) Bytes() []byte {
	var b strings.Builder
	for j, v := range i {
		b.WriteString(intString(v))
		if j < len(i)-1 {
			b.WriteString(",")
		}
	}
	return []byte(b.String())
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
	if r.Header == None {
		return []byte(r.idString())
	}
	ss := []string{r.Header.String()}
	if r.Id != 0 {
		ss = append(ss, r.idString())
	}
	if r.Data != nil {
		ss = append(ss, string(r.Data.Bytes()))
	}
	joined := strings.Join(ss, ",")

	return []byte(joined)
}

type (
	TrayStatusRequest      struct{}
	InjectionStatusRequest struct{}
)

var (
	DoInjection = &Request{
		Name:   "StartInjection",
		Header: None,
		Id:     1,
	}
	Connect = &Request{
		Name:   "Connect",
		Header: None,
		Id:     2,
	}
	Close = &Request{
		Name:   "Close",
		Header: None,
		Id:     3,
	}
	DeviceID = &Request{
		Name:   "DeviceID",
		Header: Get,
		Id:     10,
	}
	TrayStatus = &Request{
		Name:   "TrayStatus",
		Header: Get,
		Id:     12,
	}
	InjectionStatus = &Request{
		Name:   "InjectionStatus",
		Header: Get,
		Id:     13,
	}
	ExtControl = &Request{
		Name:   "ExtControl",
		Header: ExternalControl,
	}
	StartInjection = &Request{
		Name:   "StartInjection",
		Header: Inject,
	}
)

func Wait(duration int) *Request {
	return &Request{
		Name:   "Wait",
		Header: Set,
		Id:     33,
		Data:   IntData(duration),
	}
}

func Vial(position int) *Request {
	return &Request{
		Name:   "Vial",
		Header: Set,
		Id:     30,
		Data:   IntData(position),
	}
}

func AirCushion(volume int) *Request {
	return &Request{
		Name:   "AirCushion",
		Header: Set,
		Id:     21,
		Data:   IntData(volume),
	}
}

func ExcessVolume(volume int) *Request {
	return &Request{
		Name:   "ExcessVolume",
		Header: Set,
		Id:     20,
		Data:   IntData(volume),
	}
}

func NeedleDepth(depth int) *Request {
	return &Request{
		Name:   "NeedleDepth",
		Header: Set,
		Id:     26,
		Data:   IntData(depth),
	}
}

func FlushVolume(volume int) *Request {
	return &Request{
		Name:   "FlushVolume",
		Header: Set,
		Id:     22,
		Data:   IntData(volume),
	}
}

func InjectionVolume(volume int) *Request {
	return &Request{
		Name:   "InjectionVolume",
		Header: Set,
		Id:     33,
		Data:   IntData(volume),
	}
}

func InitializeCommands() []*Request {
	return []*Request{
		DeviceID,
		Connect,
		{Name: "29", Header: Set, Id: 29, Data: IntData(1)},
		Wait(2500),
		Wait(1000),
		Wait(500),
		Wait(250),
		Wait(100),
		Wait(50),
		{Name: "17", Header: Get, Id: 17},
		ExtControl,
	}
}

func HeartbeatMsgs() []*Request {
	return []*Request{
		TrayStatus,
		{Name: "17", Header: Get, Id: 17},
		InjectionStatus,
	}
}

func InjectionSettings(r *InjectRequest) ([]*Request, error) {
	target, err := vialFromGrid(r.Position, 10, 10)
	if err != nil {
		return nil, err
	}
	return []*Request{
		Vial(int(target)),
		AirCushion(int(r.AirCushion)),
		ExcessVolume(int(r.ExcessVolume)),
		{"23", Set, 23, IntData(1)},
		{"28", Set, 28, IntArrayData([]int{0, 0})},
		NeedleDepth(int(r.NeedleDepth)),
		FlushVolume(int(r.FlushVolume)),
		{"24", Set, 24, IntData(3)},
		{"90", Set, 90, IntData(0)},
		{"91", Set, 91, IntData(0)},
		{"34", Set, 34, IntData(1)},
		{"35", Set, 35, IntArrayData([]int{0, 0})},
		{"29", Set, 29, IntData(0)},
		{"27", Set, 27, IntData(0)},
		InjectionVolume(int(r.InjectionVolume)),
	}, nil
}
