package examples

import (
	"fmt"
	"github.com/jt05610/petri"
	"time"
)

func Valve() *petri.Net {

	// places
	pp := []*petri.Place{
		petri.NewPlace("A", 1, petri.Signal()),
		petri.NewPlace("B", 0, petri.Signal()),
	}

	// transitions
	tt := []*petri.Transition{
		petri.NewTransition("A Opened"),
		petri.NewTransition("B Opened"),
	}

	aa := []*petri.Arc{
		{Src: pp[1], Dest: tt[0]},
		{Src: tt[0], Dest: pp[0]},
		{Src: pp[0], Dest: tt[1]},
		{Src: tt[1], Dest: pp[1]},
	}
	return petri.NewNet("test").WithPlaces(pp...).WithTransitions(tt...).WithArcs(aa...)
}

func ExampleNet() {
	pump := NewPump()
	settings := &PumpParams{
		Flow:   1,
		Volume: 1,
	}
	fmt.Println("Initial marking with a token in the Idle place")

	fmt.Println(pump.Marking)
	var err error
	pump.Marking, err = pump.Process(pump.Marking, petri.Event[any]{
		Name: "Prepare",
		Data: settings,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("After prepare with pump parameters set to 1 flow and 1 volume")
	fmt.Println(pump.Marking)

	pump.Marking, err = pump.Process(pump.Marking, petri.Event[any]{
		Name: "Start",
		Data: 1,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Pump is pumping")
	fmt.Println(pump.Marking)
	time.Sleep(1002 * time.Millisecond)

	fmt.Println("Parameters are gone after the pump should be done, and the pump is idle again.")
	fmt.Println(pump.Marking)

	// Output:
	// Initial marking with a token in the Idle place
	// map[Idle:[Signal(1)] Pumping:[] SetParams:[]]
	// After prepare with pump parameters set to 1 flow and 1 volume
	// map[Idle:[Signal(1)] Pumping:[] SetParams:[PumpParams(&{1 1})]]
	// Pump is pumping
	// map[Idle:[] Pumping:[PumpParams(&{1 1})] SetParams:[]]
	// Parameters are gone after the pump should be done, and the pump is idle again.
	// map[Idle:[Signal(1)] Pumping:[] SetParams:[]]
}
