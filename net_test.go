package petri_test

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
)

// ExampleNet demonstrates how to create a simple net, initialize the marking, and process the marking with the net
func ExampleNet() {

	// describe the schema of the tokens involved in the system

	// we need to pay a coin to get a cookie
	coinSchema := petri.TokenSchema[interface{}]{
		Name: "Coin",
		Type: petri.Obj,
		Properties: map[string]petri.Properties{
			"Value": {
				Type: petri.Float,
			},
			"Currency": {
				Type: petri.Str,
			},
			"Diameter": {
				Type: petri.Float,
			},
		},
	}

	// we get a cookie in return
	cookieSchema := petri.TokenSchema[interface{}]{
		Name: "Cookie",
		Type: petri.Obj,
		Properties: map[string]petri.Properties{
			"Flavor": {
				Type: petri.Str,
			},
		},
	}

	signalSchema := petri.Signal()

	// the coin slot is a place where we can put coins
	coinSlot := petri.NewPlace("Coin Slot", 1, &coinSchema)

	// compartment is the place where the cookies are dispensed
	compartment := petri.NewPlace("Compartment", 1, &cookieSchema)

	// signalPlace is a place that is used to signal that a cookie is ready to be dispensed
	signalPlace := petri.NewPlace("Signal Place", 1, signalSchema)

	countSchema := petri.TokenSchema[interface{}]{
		Name: "Count",
		Type: petri.Int,
	}

	// counter is a place that keeps track of how many cookies are in the compartment
	counter := petri.NewPlace("Counter", 1, &countSchema)

	// cashBox is a place where we keep the coins that have been inserted
	cashBox := petri.NewPlace("Cash Box", 5, &coinSchema)

	// storage is a place where we keep the cookies that we have to dispense

	storage := petri.NewPlace("Storage", 5, &cookieSchema)
	type Coin struct {
		Value    float64
		Currency string
		Diameter float64
	}

	type Cookie struct {
		Flavor string
	}

	insertCoin := petri.NewTransition("Insert Coin").WithEvent(petri.NewEventFunc(func(ctx context.Context, coin *Coin) (map[string]interface{}, error) {
		return map[string]interface{}{
			"Value":    coin.Value,
			"Currency": coin.Currency,
			"Diameter": coin.Diameter,
		}, nil
	}))
	checkStorage := petri.NewTransition("Check Coin", "Coin.Currency == \"EUR\" && Coin.Value == 1.0 && Count > 0")
	returnCoin := petri.NewTransition("Return Coin", "!(Coin.Currency == \"EUR\" && Coin.Value == 1.0 && Count > 0)")
	getCookie := petri.NewTransition("Get Cookie")
	takePacket := petri.NewTransition("Take Packet").WithEvent(petri.NewEventFunc(func(ctx context.Context, cookie map[string]interface{}) (*Cookie, error) {
		flavor, ok := cookie["Flavor"].(string)
		if !ok {
			return nil, fmt.Errorf("flavor not found")
		}
		result := &Cookie{
			Flavor: flavor,
		}
		fmt.Sprintln("Got cookie", result)
		return result, nil
	}))

	machine := petri.NewNet(
		"CookieMachine",
	).WithPlaces(
		coinSlot,
		counter,
		signalPlace,
		storage,
		cashBox,
		compartment,
	).WithTransitions(
		insertCoin,
		checkStorage,
		returnCoin,
		getCookie,
		takePacket,
	).WithArcs(
		petri.NewArc(insertCoin, coinSlot, "Coin", &coinSchema),
		petri.NewArc(coinSlot, checkStorage, "Coin", &coinSchema),
		petri.NewArc(coinSlot, returnCoin, "Coin", &coinSchema),
		petri.NewArc(checkStorage, counter, "Count - 1", &countSchema),
		petri.NewArc(counter, checkStorage, "Count", &countSchema),
		petri.NewArc(counter, returnCoin, "Count", &countSchema),
		petri.NewArc(returnCoin, counter, "Count", &countSchema),
		petri.NewArc(checkStorage, cashBox, "Coin", &coinSchema),
		petri.NewArc(checkStorage, signalPlace, "Signal", signalSchema),
		petri.NewArc(signalPlace, getCookie, "Signal", signalSchema),
		petri.NewArc(storage, getCookie, "Cookie", &cookieSchema),
		petri.NewArc(getCookie, compartment, "Cookie", &cookieSchema),
		petri.NewArc(compartment, takePacket, "Cookie", &cookieSchema),
	)

	// Initialize the machine marking
	marking := machine.NewMarking()

	cookies := make([]*petri.Token[interface{}], 0)
	for i := 0; i < 5; i++ {
		cookie, err := cookieSchema.NewToken(map[string]interface{}{
			"Flavor": "Chocolate Chip",
		})
		if err != nil {
			panic(err)
		}
		cookies = append(cookies, cookie)
	}

	initialCount, err := countSchema.NewToken(5)
	if err != nil {
		panic(err)
	}
	for _, initial := range []struct {
		place  *petri.Place
		tokens []*petri.Token[interface{}]
	}{
		{counter, []*petri.Token[interface{}]{initialCount}},
		{storage, cookies},
	} {
		err = marking.PlaceTokens(initial.place, initial.tokens...)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Initial marking:")
	fmt.Println(marking)

	euro := &Coin{
		Value:    1.0,
		Currency: "EUR",
		Diameter: 23.25,
	}
	// now lets place 5 more coins in the slot and get more cookies
	for i := 0; i < 5; i++ {
		ev := petri.Event[any]{
			Name: "Insert Coin",
			Data: euro,
		}
		marking, err = machine.Process(marking, ev)
		if err != nil {
			panic(err)
		}
		take := petri.Event[any]{
			Name: "Take Packet",
		}
		marking, err = machine.Process(marking, take)
		if err != nil {
			panic(err)
		}
		fmt.Println(marking)
	}

	// then on the 6th insertion we should get a coin back
	ev := petri.Event[any]{
		Name: "Insert Coin",
		Data: euro,
	}
	marking, err = machine.Process(marking, ev)
	if err != nil {
		panic(err)
	}

	// Output:
	// Initial marking:
	// map[Cash Box:[] Coin Slot:[Coin(map[Currency:EUR Diameter:23.25 Value:1])] Compartment:[] Counter:[Count(5)] Signal Place:[] Storage:[Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip])]]
	//
	// Intermediate marking:
	// map[Cash Box:[Coin(map[Currency:EUR Diameter:23.25 Value:1])] Coin Slot:[] Compartment:[] Counter:[Count(4)] Signal Place:[Signal(<nil>)] Storage:[Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip])]]
	//
	// Final marking:
	// map[Cash Box:[Coin(map[Currency:EUR Diameter:23.25 Value:1])] Coin Slot:[] Compartment:[Cookie(map[Flavor:Chocolate Chip])] Counter:[Count(4)] Signal Place:[] Storage:[Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip]) Cookie(map[Flavor:Chocolate Chip])]]
}
