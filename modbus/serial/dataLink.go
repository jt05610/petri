package serial

import (
	"io"
	"log"
	"modbus/pdu"
)

type DataLink struct {
	io.ReadWriter
}

func (d *DataLink) Send(pdu *pdu.SerialPDU) (int, error) {
	bytes := make([]byte, len(pdu.PDU.Data)+4)
	_, err := pdu.Read(bytes)
	if err != nil {
		panic(err)
	}
	log.Printf("Sending %v\n", bytes)
	return d.Write(bytes)
}

func (d *DataLink) Recv(pdu *pdu.SerialPDU) (int, error) {
	bytes := make([]byte, 256)
	n, err := d.Read(bytes)
	if err != nil {
		panic(err)
	}
	log.Printf("Received %v\n", bytes)
	return pdu.Write(bytes[:n])
}

func NewDataLink(w io.ReadWriter) *DataLink {
	return &DataLink{ReadWriter: w}
}
