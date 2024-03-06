package couch_test

import (
	"context"
	"github.com/go-kivik/kivik/v3"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/couch"
	"testing"
)

func setUp[T petri.Object, U petri.Input, V petri.Filter, W petri.Update](name string) *couch.Service[T, U, V, W] {
	cfg := couch.LoadConfig("../.env")
	client, err := kivik.New("couch", cfg.URI())
	if err != nil {
		panic(err)
	}

	err = client.DestroyDB(context.Background(), name)
	if err != nil {
		panic(err)
	}
	err = client.CreateDB(context.Background(), name)
	if err != nil {
		panic(err)
	}

	s, err := couch.Open[T, U, V, W](cfg.URI(), name)
	if err != nil {
		panic(err)
	}
	return s
}

type TestCase[T petri.Object, U petri.Input, V petri.Filter, W petri.Update] struct {
	AddInput    U
	AddCheck    func(*testing.T, T)
	Filter      V
	ListItems   int
	ListCheck   func(*testing.T, []T)
	UpdateInput W
	UpdateCheck func(*testing.T, T)
}

func RunServiceTest[T petri.Object, U petri.Input, V petri.Filter, W petri.Update](t *testing.T, s petri.Service[T, U, V, W], tc *TestCase[T, U, V, W]) {
	t.Run("Add", func(t *testing.T) {
		res, err := s.Add(context.Background(), tc.AddInput)
		if err != nil {
			t.Error(err)
		}
		tc.AddCheck(t, res)
		_, err = s.Remove(context.Background(), res.Identifier())
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Get", func(t *testing.T) {
		res, err := s.Add(context.Background(), tc.AddInput)
		if err != nil {
			t.Error(err)
		}
		retrieved, err := s.Get(context.Background(), res.Identifier())
		if err != nil {
			t.Error(err)
		}
		tc.AddCheck(t, retrieved)
		_, err = s.Remove(context.Background(), res.Identifier())
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("List", func(t *testing.T) {
		for i := 0; i < tc.ListItems; i++ {
			_, err := s.Add(context.Background(), tc.AddInput)
			if err != nil {
				t.Error(err)
			}
		}
		res, err := s.List(context.Background(), tc.Filter)
		if err != nil {
			t.Error(err)
		}
		tc.ListCheck(t, res)
		for _, item := range res {
			_, err = s.Remove(context.Background(), item.Identifier())
			if err != nil {
				t.Error(err)
			}
		}
	})
	t.Run("Update", func(t *testing.T) {
		res, err := s.Add(context.Background(), tc.AddInput)
		if err != nil {
			t.Error(err)
		}
		updated, err := s.Update(context.Background(), res.Identifier(), tc.UpdateInput)
		if err != nil {
			t.Error(err)
		}
		tc.UpdateCheck(t, updated)
		_, err = s.Remove(context.Background(), res.Identifier())
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Remove", func(t *testing.T) {
		res, err := s.Add(context.Background(), tc.AddInput)
		if err != nil {
			t.Error(err)
		}
		_, err = s.Remove(context.Background(), res.Identifier())
		if err != nil {
			t.Error(err)
		}
		_, err = s.Get(context.Background(), res.Identifier())
		if err == nil {
			t.Error("expected", petri.ErrNotFound, "got", nil)
		}
	})
}

func TestTokenService(t *testing.T) {
	s := setUp[*petri.TokenSchema, *petri.TokenSchemaInput, *petri.TokenFilter, *petri.TokenUpdate]("token_test")
	testCase := &TestCase[*petri.TokenSchema, *petri.TokenSchemaInput, *petri.TokenFilter, *petri.TokenUpdate]{
		AddInput: &petri.TokenSchemaInput{
			Name: "person",
			Type: petri.Obj,
			Properties: map[string]petri.Properties{
				"size": {
					Type: petri.Float,
				},
				"age": {
					Type: petri.Int,
				},
				"hometown": {
					Type: petri.Obj,
					Properties: map[string]petri.Properties{
						"city": {
							Type: petri.Str,
						},
						"state": {
							Type: petri.Str,
						},
					},
				},
			},
		},
		AddCheck: func(t *testing.T, schema *petri.TokenSchema) {
			if schema.Name != "person" {
				t.Error("expected", "person", "got", schema.Name)
			}
			if schema.Type != petri.Obj {
				t.Error("expected", petri.Obj, "got", schema.Type)
			}
			if schema.Properties["size"].Type != petri.Float {
				t.Error("expected", petri.Float, "got", schema.Properties["size"].Type)
			}
			if schema.Properties["age"].Type != petri.Int {
				t.Error("expected", petri.Int, "got", schema.Properties["age"].Type)
			}
			if schema.Properties["hometown"].Type != petri.Obj {
				t.Error("expected", petri.Obj, "got", schema.Properties["hometown"].Type)
			}
			if schema.Properties["hometown"].Properties["city"].Type != petri.Str {
				t.Error("expected", petri.Str, "got", schema.Properties["hometown"].Properties["city"].Type)
			}
			if schema.Properties["hometown"].Properties["state"].Type != petri.Str {
				t.Error("expected", petri.Str, "got", schema.Properties["hometown"].Properties["state"].Type)
			}
		},
		Filter: &petri.TokenFilter{
			Name: &petri.StringSelector{
				Equals: "person",
			},
		},
		ListItems: 5,
		ListCheck: func(t *testing.T, schemas []*petri.TokenSchema) {
			if len(schemas) != 5 {
				t.Error("expected", 5, "got", len(schemas))
			}
		},
		UpdateInput: &petri.TokenUpdate{
			Name: "person2",
			Mask: petri.TokenMask{
				Name: true,
			},
		},
		UpdateCheck: func(t *testing.T, schema *petri.TokenSchema) {
			if schema.Name != "person2" {
				t.Error("expected", "person2", "got", schema.Name)
			}
		},
	}
	RunServiceTest[*petri.TokenSchema, *petri.TokenSchemaInput, *petri.TokenFilter, *petri.TokenUpdate](t, s, testCase)
}

func TestPlaceService(t *testing.T) {
	fakeToken := &petri.TokenSchema{
		ID:   "fake",
		Name: "Count",
		Type: petri.Int,
	}

	s := setUp[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate]("token_test")
	testCase := &TestCase[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate]{
		AddInput: &petri.PlaceInput{
			Name: "place",
			AcceptedTokens: []*petri.TokenSchema{
				fakeToken,
			},
			Bound: 1,
		},
		AddCheck: func(t *testing.T, place *petri.Place) {
			if place.Name != "place" {
				t.Error("expected", "person", "got", place.Name)
			}
			if len(place.AcceptedTokens) != 1 {
				t.Error("expected", 1, "got", len(place.AcceptedTokens))
			}
			if place.AcceptedTokens[0].ID != fakeToken.ID {
				t.Error("expected", fakeToken.Name, "got", place.AcceptedTokens[0].Name)
			}
			if place.Bound != 1 {
				t.Error("expected", 1, "got", place.Bound)
			}
		},
		Filter: &petri.PlaceFilter{
			Bound: &petri.IntSelector{
				Equals: 1,
			},
		},
		ListItems: 5,
		ListCheck: func(t *testing.T, places []*petri.Place) {
			if len(places) != 5 {
				t.Error("expected", 5, "got", len(places))
			}
		},
		UpdateInput: &petri.PlaceUpdate{
			Input: &petri.PlaceInput{
				Name: "place2",
			},
			Mask: &petri.PlaceMask{
				Name: true,
			},
		},
		UpdateCheck: func(t *testing.T, place *petri.Place) {
			if place.Name != "place2" {
				t.Error("expected", "place2", "got", place.Name)
			}
		},
	}
	RunServiceTest[*petri.Place, *petri.PlaceInput, *petri.PlaceFilter, *petri.PlaceUpdate](t, s, testCase)
}
func TestTransitionService(t *testing.T) {
	s := setUp[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate]("token_test")
	testCase := &TestCase[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate]{
		AddInput: &petri.TransitionInput{
			Name: "transition",
		},
		AddCheck: func(t *testing.T, place *petri.Transition) {
			if place.Name != "transition" {
				t.Error("expected", "person", "got", place.Name)
			}
		},
		Filter: &petri.TransitionFilter{
			Name: &petri.StringSelector{
				Equals: "transition",
			},
		},
		ListItems: 5,
		ListCheck: func(t *testing.T, places []*petri.Transition) {
			if len(places) != 5 {
				t.Error("expected", 5, "got", len(places))
			}
		},
		UpdateInput: &petri.TransitionUpdate{
			Input: &petri.TransitionInput{
				Name: "transition2",
			},
			Mask: &petri.TransitionMask{
				Name: true,
			},
		},
		UpdateCheck: func(t *testing.T, place *petri.Transition) {
			if place.Name != "transition2" {
				t.Error("expected", "transition2", "got", place.Name)
			}
		},
	}
	RunServiceTest[*petri.Transition, *petri.TransitionInput, *petri.TransitionFilter, *petri.TransitionUpdate](t, s, testCase)
}

func TestArcService(t *testing.T) {
	signal := &petri.TokenSchema{
		ID:   "signal",
		Name: "signal",
		Type: petri.Bool,
	}
	head := petri.NewPlace("head", 1, signal)
	newHead := petri.NewPlace("newHead", 1, signal)
	tail := petri.NewTransition("arc")

	s := setUp[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate]("token_test")
	testCase := &TestCase[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate]{
		AddInput: &petri.ArcInput{
			Src:          head,
			Dest:         tail,
			OutputSchema: signal,
			Expression:   "signal",
		},
		AddCheck: func(t *testing.T, arc *petri.Arc) {
			if arc.Expression != "signal" {
				t.Error("expected", "signal", "got", arc.Expression)
			}
			if arc.Src.Identifier() != head.Identifier() {
				t.Error("expected", head.Identifier(), "got", arc.Src.Identifier())
			}
			if arc.Dest.Identifier() != tail.Identifier() {
				t.Error("expected", tail.Identifier(), "got", arc.Dest.Identifier())
			}
			if arc.OutputSchema.Identifier() != signal.Identifier() {
				t.Error("expected", signal.Identifier(), "got", arc.OutputSchema.Identifier())
			}
		},
		Filter: &petri.ArcFilter{
			Src: &petri.PlaceFilter{
				ID: &petri.StringSelector{
					Equals: head.Identifier(),
				},
			},
		},
		ListItems: 5,
		ListCheck: func(t *testing.T, arcs []*petri.Arc) {
			if len(arcs) != 5 {
				t.Error("expected", 5, "got", len(arcs))
			}
		},
		UpdateInput: &petri.ArcUpdate{
			Input: &petri.ArcInput{
				Src: newHead,
			},
			Mask: &petri.ArcMask{
				Src: true,
			},
		},
		UpdateCheck: func(t *testing.T, arc *petri.Arc) {
			if arc.Src.Identifier() != newHead.Identifier() {
				t.Error("expected", newHead.Identifier(), "got", arc.Src.Identifier())
			}
		},
	}
	RunServiceTest[*petri.Arc, *petri.ArcInput, *petri.ArcFilter, *petri.ArcUpdate](t, s, testCase)
}

func TestNetService(t *testing.T) {
	signal := &petri.TokenSchema{
		ID:   "signal",
		Name: "signal",
		Type: petri.Bool,
	}
	head := petri.NewPlace("head", 1, signal)
	tail := petri.NewTransition("net")
	net := petri.NewNet("net").WithPlaces(
		head,
	).WithTransitions(
		tail,
	).WithArcs(
		petri.NewArc(head, tail, "signal", signal),
	)
	s := setUp[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate]("token_test")
	testCase := &TestCase[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate]{
		AddInput: &petri.NetInput{
			Name:         net.Name,
			TokenSchemas: []*petri.TokenSchema{signal},
			Places:       []*petri.Place{head},
			Transitions:  []*petri.Transition{tail},
			Arcs:         []*petri.Arc{petri.NewArc(head, tail, "signal", signal)},
		},
		AddCheck: func(t *testing.T, net *petri.Net) {
			if net.Name != "net" {
				t.Error("expected", "net", "got", net.Name)
			}
			if len(net.TokenSchemas) != 1 {
				t.Error("expected", 1, "got", len(net.TokenSchemas))
			}
			if len(net.Places) != 1 {
				t.Error("expected", 1, "got", len(net.Places))
			}
			if len(net.Transitions) != 1 {
				t.Error("expected", 1, "got", len(net.Transitions))
			}
			if len(net.Arcs) != 1 {
				t.Error("expected", 1, "got", len(net.Arcs))
			}
			if net.TokenSchemas[0].Identifier() != signal.Identifier() {
				t.Error("expected", signal.Identifier(), "got", net.TokenSchemas[0].Identifier())
			}
			if net.Places[0].Identifier() != head.Identifier() {
				t.Error("expected", head.Identifier(), "got", net.Places[0].Identifier())
			}
			if net.Transitions[0].Identifier() != tail.Identifier() {
				t.Error("expected", tail.Identifier(), "got", net.Transitions[0].Identifier())
			}
			if net.Arcs[0].Identifier() != net.Arcs[0].Identifier() {
				t.Error("expected", net.Arcs[0].Identifier(), "got", net.Arcs[0].Identifier())
			}
		},
		Filter: &petri.NetFilter{
			Name: &petri.StringSelector{
				Equals: net.Name,
			},
		},
		ListItems: 5,
		ListCheck: func(t *testing.T, nets []*petri.Net) {
			if len(nets) != 5 {
				t.Error("expected", 5, "got", len(nets))
			}
		},
		UpdateInput: &petri.NetUpdate{
			Input: &petri.NetInput{
				Name: "net2",
			},
			Mask: &petri.NetMask{
				Name: true,
			},
		},
		UpdateCheck: func(t *testing.T, net *petri.Net) {
			if net.Name != "net2" {
				t.Error("expected", "net2", "got", net.Name)
			}
		},
	}
	RunServiceTest[*petri.Net, *petri.NetInput, *petri.NetFilter, *petri.NetUpdate](t, s, testCase)
}
