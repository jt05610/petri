package pdu_test

import (
	pdu2 "modbus/pdu"
	"testing"
)

func TestNewSerialPDU(t *testing.T) {
	mPDU := pdu2.ReadCoils(0xFEED, 0xBEAD)
	sPDU := pdu2.NewSerialPDU(0x01, mPDU)
	expected := []byte{0x01, 0x01, 0xFE, 0xED, 0xBE, 0xAD, 0x2D, 0xCA}
	actual := make([]byte, len(expected))
	n, err := sPDU.Read(actual)
	if err != nil {
		t.Error(err)
	}
	if n != len(expected) {
		t.Logf("expected %v bytes but got %v", len(expected), n)
		t.Fail()
	}
	for i := 0; i < n; i++ {
		if expected[i] != actual[i] {
			t.Logf("mismatch at %v: expected %v but got %v", i, expected[i], actual[i])
			t.Fail()
		}
	}

	actualPDU := &pdu2.SerialPDU{}
	_, err = actualPDU.Write(expected)
	if err != nil {
		t.Error(err)
	}
	if sPDU.Addr != actualPDU.Addr {
		t.Fail()
	}
	if sPDU.CRC != actualPDU.CRC {
		t.Fail()
	}
}

func BenchmarkNewSerialPDU(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mPDU := pdu2.ReadCoils(0xFEED, 0xBEAD)
		pdu2.NewSerialPDU(0x01, mPDU)
	}
}
