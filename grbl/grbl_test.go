package grbl_test

import (
	"bytes"
	"github.com/jt05610/petri/grbl"
	"testing"
)

var testCases = []struct {
	name   string
	buffer []byte
	expect grbl.StatusUpdate
}{
	{
		name:   "workCoordinate",
		buffer: []byte("<Alarm|MPos:0.000,0.000,0.000|F:0|WCO:0.000,0.000,-16.275>\n"),
		expect: &grbl.Status{
			Alarm: grbl.Alarm,
			MachinePosition: &grbl.Position{
				X: 0,
				Y: 0,
				Z: 0,
			},
			Feed: 0,
			WorkPosition: &grbl.Position{
				X: 0,
				Y: 0,
				Z: -16.275,
			},
		},
	},
	{
		name:   "ok",
		buffer: []byte("ok\n"),
		expect: &grbl.Ack{},
	},
	{
		name:   "0v",
		buffer: []byte("<Alarm|MPos:0.000,0.000,0.000|F:0|Ov:100,100,100>\n"),
		expect: &grbl.Status{
			Alarm: grbl.Alarm,
			MachinePosition: &grbl.Position{
				X: 0,
				Y: 0,
				Z: 0,
			},
			Feed: 0,
			Override: &struct {
				X float64
				Y float64
				Z float64
			}{
				X: 100,
				Y: 100,
				Z: 100,
			},
		},
	},
	{
		name:   "F",
		buffer: []byte("<Alarm|MPos:0.000,0.000,0.000|F:0>\n"),
		expect: &grbl.Status{
			Alarm: grbl.Alarm,
			MachinePosition: &grbl.Position{
				X: 0,
				Y: 0,
				Z: 0,
			},
			Feed: 0,
		},
	},
}

func TestParse(t *testing.T) {
	for _, tc := range testCases {
		p := grbl.NewParser(bytes.NewReader(tc.buffer))
		u, err := p.Parse()
		if err != nil {
			t.Fatal(err)
		}
		if u == nil {
			t.Fatal("nil status")
		}
		if s, ok := u.(*grbl.Status); ok {
			if e, ok := tc.expect.(*grbl.Status); ok {
				if s.Alarm != e.Alarm {
					t.Fatalf("expected alarm %v, got %v", e.Alarm, s.Alarm)
				}
				if s.Feed != e.Feed {
					t.Fatalf("expected feed %v, got %v", e.Feed, s.Feed)
				}
				if s.MachinePosition.X != e.MachinePosition.X {
					t.Fatalf("expected machine position x %v, got %v", e.MachinePosition.X, s.MachinePosition.X)
				}
				if s.MachinePosition.Y != e.MachinePosition.Y {
					t.Fatalf("expected machine position y %v, got %v", e.MachinePosition.Y, s.MachinePosition.Y)
				}
				if s.MachinePosition.Z != e.MachinePosition.Z {
					t.Fatalf("expected machine position z %v, got %v", e.MachinePosition.Z, s.MachinePosition.Z)
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
			if _, ok := tc.expect.(*grbl.Ack); !ok {
				t.Fatalf("expected ack, got %T", tc.expect)
			}
		}
	}
}
