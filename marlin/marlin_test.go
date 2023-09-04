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
		buffer: []byte("X:0.00 Y:0.00 Z:0.00 E:0.00 Count X:0 Y:0 Z:0"),
		expect: &marlin.Status{
			Alarm: nil,
			State: "alarm",
			Position: &marlin.Position{
				X: 0,
				Y: 0,
				Z: 0,
			},
			Feed: 0,
			WorkPosition: &marlin.Position{
				X: 0,
				Y: 0,
				Z: -16.275,
			},
		},
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
				if s.Feed != e.Feed {
					t.Fatalf("expected feed %v, got %v", e.Feed, s.Feed)
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
				if s.WorkPosition != nil {
					if e.WorkPosition == nil {
						t.Fatalf("expected work position nil, got %v", s.WorkPosition)
					}
					if s.WorkPosition.X != e.WorkPosition.X {
						t.Fatalf("expected work position x %v, got %v", e.WorkPosition.X, s.WorkPosition.X)
					}
					if s.WorkPosition.Y != e.WorkPosition.Y {
						t.Fatalf("expected work position y %v, got %v", e.WorkPosition.Y, s.WorkPosition.Y)
					}
					if s.WorkPosition.Z != e.WorkPosition.Z {
						t.Fatalf("expected work position z %v, got %v", e.WorkPosition.Z, s.WorkPosition.Z)
					}
				}
				if s.Override != nil {
					if e.Override == nil {
						t.Fatalf("expected override nil, got %v", s.Override)
					}
					if s.Override.Rapid != e.Override.Rapid {
						t.Fatalf("expected override x %v, got %v", e.Override.Rapid, s.Override.Rapid)
					}
					if s.Override.Feed != e.Override.Feed {
						t.Fatalf("expected override y %v, got %v", e.Override.Feed, s.Override.Feed)
					}
					if s.Override.Spindle != e.Override.Spindle {
						t.Fatalf("expected override z %v, got %v", e.Override.Spindle, s.Override.Spindle)
					}
				}

			} else {
				t.Fatalf("expected status, got %T", tc.expect)
			}

		} else {
			if _, ok := tc.expect.(*marlin.Ack); !ok {
				t.Fatalf("expected ack, got %T", tc.expect)
			}
		}
	}
}
