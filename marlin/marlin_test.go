package marlin_test

import (
	"bytes"
	"github.com/jt05610/petri/marlin"
	"testing"
)

var testCases = []struct {
	name   string
	buffer []byte
	expect marlin.StatusUpdate
}{
	{
		name:   "pos upd",
		buffer: []byte("X:0.00 Y:0.00 Z:40.00 E:0.00 Count X:0 Y:0 Z:16000"),
		expect: &marlin.Status{
			Alarm: nil,
			Position: &marlin.Position{
				X: 0,
				Y: 0,
				Z: 40,
			},
		},
	},
	{
		name:   "busy processing",
		buffer: []byte("echo:busy: processing"),
		expect: &marlin.Processing{},
	},
	{
		name:   "ok",
		buffer: []byte("ok"),
		expect: &marlin.Ack{},
	},
}

func TestParse(t *testing.T) {
	for _, tc := range testCases {
		p := marlin.NewParser(bytes.NewReader(tc.buffer))
		u, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		if u == nil {
			t.Fatal("nil status")
		}
		if s, ok := u.(*marlin.Status); ok {
			if e, ok := tc.expect.(*marlin.Status); ok {
				if s.Alarm != e.Alarm {
					t.Fatalf("expected alarm %v, got %v", e.Alarm, s.Alarm)
				}
				if s.Position.X != e.Position.X {
					t.Fatalf("expected machine position x %v, got %v", e.Position.X, s.Position.X)
				}
				if s.Position.Y != e.Position.Y {
					t.Fatalf("expected machine position y %v, got %v", e.Position.Y, s.Position.Y)
				}
				if s.Position.Z != e.Position.Z {
					t.Fatalf("expected machine position z %v, got %v", e.Position.Z, s.Position.Z)
				}

			} else {
				t.Fatalf("expected status, got %T", tc.expect)
			}

		}
		if _, ok := u.(*marlin.Ack); ok {
			if _, ok := tc.expect.(*marlin.Ack); !ok {
				t.Fatalf("expected ack, got %T", tc.expect)
			}
		}
		if _, ok := u.(*marlin.Processing); ok {
			if _, ok := tc.expect.(*marlin.Processing); !ok {
				t.Fatalf("expected processing, got %T", tc.expect)
			}
		}
	}
}
