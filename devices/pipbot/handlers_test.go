package PipBot_test

import (
	PipBot "github.com/jt05610/petri/devices/pipbot"
	"testing"
)

type TestCase struct {
	name          string
	initialState  *PipBot.State
	expectedState *PipBot.State
	plan          *PipBot.TransferPlan
	nMoves        int
}

func TestTransferPlan_Moves(t *testing.T) {
	pb := PipBot.NewPipBot([]int{3, 4}, 0, nil)
	initial := pb.State()
	hasTipWithFluidState := &PipBot.State{
		Pipette: &PipBot.Fluid{
			Source: &PipBot.Location{
				Grid: 0,
				Row:  0,
				Col:  0,
			},
			Volume: 100,
		},
		HasTip:     true,
		TipChannel: initial.TipChannel,
		Layout:     initial.Layout,
		TipIndex:   0,
	}
	hasTipNoFluidState := &PipBot.State{
		Pipette: &PipBot.Fluid{
			Source: &PipBot.Location{
				Grid: 0,
				Row:  0,
				Col:  0,
			},
			Volume: 0,
		},
		HasTip:     true,
		TipChannel: initial.TipChannel,
		Layout:     initial.Layout,
		TipIndex:   0,
	}

	cases := []*TestCase{
		{
			name:         "NoTip",
			nMoves:       17,
			initialState: initial,
			plan: &PipBot.TransferPlan{
				Technique: PipBot.Forward,
				Source: &PipBot.Location{
					Grid: 0,
					Row:  0,
					Col:  0,
				},
				Dest: &PipBot.Location{
					Grid: 0,
					Row:  0,
					Col:  1,
				},
				ReplaceTip:     true,
				PreRinses:      2,
				FillVolume:     100,
				ExcessVolume:   0,
				AspirationRate: 5,
				DispenseVolume: 50,
				DispenseRate:   50,
				DwellTime:      1000,
			},
			expectedState: &PipBot.State{
				Pipette: &PipBot.Fluid{
					Source: &PipBot.Location{
						Grid: 0,
						Row:  0,
						Col:  0,
					},
					Volume: 50,
				},
				HasTip: true,
			},
		},
		{
			name:         "HasTipNoFluidReplace",
			initialState: hasTipNoFluidState,
			expectedState: &PipBot.State{
				Pipette: &PipBot.Fluid{
					Source: &PipBot.Location{
						Grid: 0,
						Row:  0,
						Col:  0,
					},
					Volume: 50,
				},
				HasTip: true,
			},
		},
		{
			name:         "HasTipWithFluidMulti",
			initialState: hasTipWithFluidState,
			expectedState: &PipBot.State{
				Pipette: &PipBot.Fluid{
					Source: &PipBot.Location{
						Grid: 0,
						Row:  0,
						Col:  0,
					},
					Volume: 50,
				},
				HasTip: true,
			},
		},
		{
			name:         "HasTipWithFluidReplace",
			initialState: hasTipWithFluidState,
			expectedState: &PipBot.State{
				Pipette: &PipBot.Fluid{
					Source: &PipBot.Location{
						Grid: 0,
						Row:  0,
						Col:  0,
					},
					Volume: 50,
				},
				HasTip: true,
			},
		},
		{
			name:         "HasTipWithFluidNoReplace",
			initialState: hasTipWithFluidState,
			expectedState: &PipBot.State{
				Pipette: &PipBot.Fluid{
					Source: &PipBot.Location{
						Grid: 0,
						Row:  0,
						Col:  0,
					},
					Volume: 50,
				},
				HasTip: true,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := &PipBot.State{
				Pipette:    c.initialState.Pipette,
				HasTip:     c.initialState.HasTip,
				TipChannel: c.initialState.TipChannel,
				Layout:     c.initialState.Layout,
				TipIndex:   c.initialState.TipIndex,
			}
			newState, moves, err := c.plan.Moves(state)
			if err != nil {
				t.Errorf("error generating moves: %s", err)
			}
			if len(moves) != c.nMoves {
				t.Errorf("expected %d moves, got %d", c.nMoves, len(moves))
			}
			if newState.Pipette != nil {
				if c.expectedState.Pipette == nil {
					t.Errorf("expected pipette to be nil, got %v", newState.Pipette)
				}
				if newState.Pipette.Source != nil {
					if c.expectedState.Pipette.Source == nil {
						t.Errorf("expected pipette source to be nil, got %v", newState.Pipette.Source)
					}
					if !c.expectedState.Pipette.Source.Equals(newState.Pipette.Source) {
						t.Errorf("expected pipette source to be %v, got %v", c.expectedState.Pipette.Source, newState.Pipette.Source)
					}
				}
				if newState.Pipette.Volume != c.expectedState.Pipette.Volume {
					t.Errorf("expected pipette volume to be %f, got %f", c.expectedState.Pipette.Volume, newState.Pipette.Volume)
				}
			}
			if newState.Pipette == nil {
				if c.expectedState.Pipette != nil {
					t.Errorf("expected pipette to be %v, got nil", c.expectedState.Pipette)
				}
			}
		})
	}
}
