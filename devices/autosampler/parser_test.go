package autosampler

import (
	"reflect"
	"strings"
	"testing"
)

type testCase struct {
	name     string
	input    string
	expected []*Response[Byteable]
}

func TestParser_Parse(t *testing.T) {
	cases := []testCase{
		{"DeviceID", "G,10,Series 200 Autosamp\r",
			[]*Response[Byteable]{
				{
					Request: DeviceID,
					Data:    StringData("Series 200 Autosamp"),
				},
			},
		},
		{"Connect", "0\r",
			[]*Response[Byteable]{
				{
					Data: IntData(0),
				},
			},
		},
		{
			"TrayStatus", "G,12,1\r",
			[]*Response[Byteable]{
				{
					Request: TrayStatus,
					Data:    IntData(1),
				},
			},
		},
		{
			"InjectBeforeStatus", "I\nG,13,3\r",
			[]*Response[Byteable]{
				{
					Request: &Request{Header: Inject},
				},
				{
					Request: InjectionStatus,
					Data:    IntData(3),
				},
			},
		},
		{
			"InjectAfterStatus", "G,13,3\rI\n",
			[]*Response[Byteable]{
				{
					Request: &Request{Header: Inject},
				},
				{
					Request: InjectionStatus,
					Data:    IntData(3),
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(c.input))
			gotArray, err := p.Parse()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(gotArray) != len(c.expected) {
				t.Errorf("expected %d responses, got %d", len(c.expected), len(gotArray))
			}
			for i, e := range c.expected {
				got := gotArray[i]
				if e.Request != nil {
					if got.Request == nil {
						t.Errorf("expected %v, got %v", e.Request, got.Request)
					}
					if !reflect.DeepEqual(got.Request.Bytes(), e.Request.Bytes()) {
						t.Errorf("expected %v, got %v", e.Request, got.Request)
					}
				}
				if e.Data != nil {
					if got.Data == nil {
						t.Errorf("expected %v, got %v", e.Data, got.Data)
					}
					if !reflect.DeepEqual(got.Data.Bytes(), e.Data.Bytes()) {
						t.Errorf("expected %v, got %v", e.Data, got.Data)
					}
				}
			}
		})
	}
}
