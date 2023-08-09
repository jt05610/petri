package serial

import (
	"errors"
	"golang.org/x/net/context"
	"modbus/pdu"
	"modbus/wire"
	"sync"
)

type Client struct {
	*DataLink
	mu sync.Mutex
}

func (c *Client) Request(ctx context.Context, addr byte, req *pdu.ModbusPDU) (*pdu.ModbusPDU, error) {
	res := &pdu.SerialPDU{}
	s := pdu.NewSerialPDU(addr, req)
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.DataLink.Send(s)
	_, err = c.DataLink.Recv(res)
	if err != nil {
		panic(err)
	}
	if res.Addr != addr {
		return nil, errors.New("invalid response address")
	}
	return res.PDU, nil
}

func DefaultClient() *Client {
	ser, err := wire.NewSerial()
	if err != nil {
		panic(err)
	}
	return &Client{
		DataLink: NewDataLink(ser),
	}
}
