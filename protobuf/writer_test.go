package protobuf

import (
	"github.com/jt05610/petri/examples"
	"strings"
	"testing"
)

func TestService_Flush(t *testing.T) {
	net := examples.Switch()

	s := Service{}
	wr := &strings.Builder{}
	err := s.Flush(nil, wr, net)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(wr.String())
}
