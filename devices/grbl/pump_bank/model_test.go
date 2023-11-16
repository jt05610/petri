package pump_bank

import (
	proto "github.com/jt05610/petri/grbl/proto/v1"
	"math"
	"testing"
)

func TestAqOrgRate(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		input                   *StartPumpRequest
		expectedAq, expectedOrg float64
	}{
		{
			name: "equal",
			input: &StartPumpRequest{
				Volume: 10,
				TFR:    1,
				FRR:    1,
			},
			expectedAq:  .5,
			expectedOrg: .5,
		},
		{
			name: "2x_aqueous",
			input: &StartPumpRequest{
				Volume: 10,
				TFR:    1,
				FRR:    2,
			},
			expectedAq:  2.0 / 3.0,
			expectedOrg: 1.0 / 3.0,
		},
		{
			name: "2x_organic",
			input: &StartPumpRequest{
				Volume: 10,
				TFR:    1,
				FRR:    .5,
			},
			expectedAq:  1.0 / 3.0,
			expectedOrg: 2.0 / 3.0,
		},
		{
			name: "3-to-1",
			input: &StartPumpRequest{
				Volume: -5,
				TFR:    5,
				FRR:    3,
			},
			expectedAq:  (3.0 / 4.0) * 5,
			expectedOrg: (1.0 / 4.0) * 5,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualAq, actualOrg := tc.input.aqueousOrganicRate()
			if !floatEquals(actualAq, tc.expectedAq) {
				t.Errorf("expected aq %f, got %f", tc.expectedAq, actualAq)
			}
			if !floatEquals(actualOrg, tc.expectedOrg) {
				t.Errorf("expected org %f, got %f", tc.expectedOrg, actualOrg)
			}
		})
	}
}

func floatEquals(aq float64, aq2 float64) bool {
	return math.Abs(aq-aq2) < 0.0001
}

func TestMakeRequest(t *testing.T) {
	b := NewPumpBank(nil)
	err := b.Initialize(nil, &BankSettings{
		Aqueous: &Settings{
			SyringeDiameter: 26.7,
			SyringeVolume:   60,
			MaxDistance:     50,
			MaxFeedRate:     50,
			currentPosition: -6.697595,
		},
		Organic: &Settings{
			SyringeDiameter: 12.06,
			SyringeVolume:   5,
			MaxDistance:     50,
			MaxFeedRate:     50,
			currentPosition: -10.94273,
		},
	})
	if err != nil {
		t.Error(err)
	}
	for _, tc := range []struct {
		name     string
		input    *StartPumpRequest
		expected *proto.MoveRequest
	}{
		{
			name: "happy_params",
			input: &StartPumpRequest{
				Volume: 5,
				TFR:    5,
				FRR:    3,
			},
			expected: makeExpected(0, 0, 12.82969),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := b.makeRequest(tc.input)
			if err != nil {
				t.Error(err)
			}
			if !floatEquals(float64(actual.GetX()), float64(tc.expected.GetX())) {
				t.Errorf("expected x %f, got %f", tc.expected.GetX(), actual.GetX())
			}
			if !floatEquals(float64(actual.GetY()), float64(tc.expected.GetY())) {
				t.Errorf("expected y %f, got %f", tc.expected.GetY(), actual.GetY())
			}
			if !floatEquals(float64(actual.GetSpeed()), float64(tc.expected.GetSpeed())) {
				t.Errorf("expected speed %f, got %f", tc.expected.GetSpeed(), actual.GetSpeed())
			}
		})
	}
}

func makeExpected(x, y, rate float32) *proto.MoveRequest {
	return &proto.MoveRequest{
		X:     &x,
		Y:     &y,
		Speed: &rate,
	}
}
