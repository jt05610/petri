package petri_test

import (
	"errors"
	"fmt"
	"github.com/jt05610/petri"
)

// ExampleNet demonstrates how to create a simple net, initialize the marking, and process the marking with the net
func ExampleNet() {

	// describe the schema of the tokens involved in the system

	// we need to pay a coin to get a cookie
	coinSchema := petri.TokenSchema{
		Name: "Coin",
		Type: petri.Obj(petri.Properties{
			"Value":    petri.Float,
			"Currency": petri.String,
			"Diameter": petri.Float,
		}),
	}

	// we get a cookie in return
	cookieSchema := petri.TokenSchema{
		Name: "Cookie",
		Type: petri.Obj(petri.Properties{
			"Flavor": petri.String,
		}),
	}

	// the coin slot is a place where we can put coins
	coinSlot := petri.Place{
		Name:           "Coin Slot",
		Bound:          1,
		AcceptedTokens: []*petri.TokenSchema{&coinSchema},
	}

	// compartment is the place where the cookies are dispensed
	compartment := petri.Place{
		Name:           "Compartment",
		Bound:          1,
		AcceptedTokens: []*petri.TokenSchema{&cookieSchema},
	}

	// t is the transition that takes in a coin and dispenses a cookie
	t := petri.NewTransition("t").WithTransformer(func(input *petri.Token) (*petri.Token, error) {
		// make sure it is a coin
		if input.Schema.Name != "Coin" {
			return nil, errors.New("not a coin")
		}
		if !input.IsValid() {
			return nil, errors.New("invalid coin")
		}
		coinProps := input.Value.(map[string]interface{})
		// make sure it is a euro coin
		if coinProps["Currency"].(string) != "EUR" {
			return nil, errors.New("not a euro coin")
		}
		// make sure it is a 1 euro coin
		if coinProps["Value"].(float64) != 1.0 {
			return nil, errors.New("not a 1 euro coin")
		}
		// everything checks out, enjoy your cookie
		return &petri.Token{
			Schema: &cookieSchema,
			Value: map[string]interface{}{
				"Flavor": "Chocolate Chip",
			},
		}, nil
	})

	machine := petri.NewNet(
		"CookieMachine",
	).WithPlaces(
		&coinSlot,
		&compartment,
	).WithTransitions(
		t,
	).WithArcs(
		petri.NewArc(&coinSlot, t),
		petri.NewArc(t, &compartment),
	)

	// Initialize the machine marking
	marking := machine.NewMarking()

	// make a 1 euro coin to put in the machine
	euro, err := coinSchema.NewToken(map[string]interface{}{
		"Value":    1.0,
		"Currency": "EUR",
		"Diameter": 23.25,
	})
	if err != nil {
		panic(err)
	}

	// update the marking with the euro
	err = marking.PlaceTokens(&coinSlot, euro)
	if err != nil {
		panic(err)
	}

	fmt.Println("Initial marking:")
	fmt.Println(marking)

	// now if we process the marking with the cookie machine we should update the marking
	// and have a cookie in the compartment and no coins in the coin slot
	marking, err = machine.Process(marking)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nFinal marking:")
	fmt.Println(marking)

	// Output:
	// Initial marking:
	// map[Coin Slot:[Coin(map[Currency:EUR Diameter:23.25 Value:1])] Compartment:[]]
	//
	// Final marking:
	// map[Coin Slot:[] Compartment:[Cookie(map[Flavor:Chocolate Chip])]]
}
