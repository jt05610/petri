package examples

import (
	"context"
	"github.com/jt05610/petri"
	"time"
)

type PumpParams struct {
	Flow   float64 `json:"flow"`
	Volume float64 `json:"volume"`
}

// Pump is a simple pump that pumps a volume of liquid at a given flow rate
type Pump struct {
	Settings *PumpParams
	petri.Marking
	*petri.Net
	startedAt time.Time
	pumping   bool
}

func (p *Pump) Stop(ctx context.Context, input interface{}) (int, error) {
	p.pumping = false
	return 0, nil
}

func (p *Pump) Prepare(ctx context.Context, input *PumpParams) (*PumpParams, error) {
	p.Settings = input
	return input, nil
}

func (p *Pump) Start(ctx context.Context, input interface{}) (*PumpParams, error) {
	var err error
	p.startedAt = time.Now()
	p.pumping = true
	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(1e9*p.Settings.Volume/p.Settings.Flow) * time.Nanosecond):
			p.Marking, err = p.Net.Process(p.Marking, petri.Event[any]{
				Name: "Stop",
				Data: 1,
			})
			if err != nil {
				panic(err)
			}
		}
	}()
	return p.Settings, nil
}

func NewPump() *Pump {

	signal := petri.Signal()

	pumpParams := &petri.TokenSchema{
		ID:   petri.ID(),
		Name: "PumpParams",
		Type: "object",
		Properties: map[string]petri.Properties{
			"flow": {
				Type: "number",
			},
			"volume": {
				Type: "number",
			},
		},
	}

	net := petri.NewNet("Pump")

	// struct that the pumpParameters schema represents

	// Initializing the pump in the idle state
	tok, err := signal.NewToken(1)
	if err != nil {
		panic(err)
	}

	pp := []*petri.Place{
		petri.NewPlace("Idle", 1, petri.Signal()),
		petri.NewPlace("SetParams", 1, pumpParams),
		petri.NewPlace("Pumping", 1, pumpParams),
	}

	p := &Pump{
		Net: net.WithPlaces(pp...),
	}

	p.Marking = p.NewMarking()

	err = p.Marking.PlaceTokens(p.Net.Place("Idle"), tok)
	if err != nil {
		panic(err)

	}

	tt := []*petri.Transition{
		petri.NewTransition("Start").WithEvent(petri.NewEventFunc(p.Start), "http://localhost:8080/pump/start", petri.Signal(), petri.Signal()),
		petri.NewTransition("Stop").WithEvent(petri.NewEventFunc(p.Stop), "http://localhost:8080/pump/stop", petri.Signal(), petri.Signal()),
		petri.NewTransition("Prepare").WithEvent(petri.NewEventFunc(p.Prepare), "http://localhost:8080/pump/prepare", pumpParams, pumpParams),
	}

	p.Net = p.Net.WithTransitions(tt...)

	aa := []*petri.Arc{
		petri.NewArc(p.Net.Place("Idle"), p.Net.Transition("Start"), "Signal", signal),
		petri.NewArc(p.Net.Transition("Prepare"), p.Net.Place("SetParams"), "PumpParams", pumpParams),
		petri.NewArc(p.Net.Place("SetParams"), p.Net.Transition("Start"), "PumpParams", pumpParams),
		petri.NewArc(p.Net.Transition("Start"), p.Net.Place("Pumping"), "PumpParams", pumpParams),
		petri.NewArc(p.Net.Place("Pumping"), p.Net.Transition("Stop"), "PumpParams", pumpParams),
		petri.NewArc(p.Net.Transition("Stop"), p.Net.Place("Idle"), "Signal", signal),
	}

	p.Net = p.Net.WithArcs(aa...)
	return p
}
