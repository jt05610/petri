package serial

import "testing"

func TestListPorts(t *testing.T) {
	pp, err := ListPorts()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pp {
		t.Log(p)
	}
}
