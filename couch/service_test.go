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

func TestService_Add(t *testing.T) {
	s := setUp[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate]("token_test")
	in := &petri.TokenInput{
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
						Type: petri.String,
					},
					"state": {
						Type: petri.String,
					},
				},
			},
		},
	}
	schema, err := s.Add(context.Background(), in)
	if err != nil {
		t.Error(err)
	}
	if schema.Name != in.Name {
		t.Error("expected", in.Name, "got", schema.Name)
	}
}

func makePerson(s *couch.Service[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate]) *petri.TokenSchema {
	in := &petri.TokenInput{
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
						Type: petri.String,
					},
					"state": {
						Type: petri.String,
					},
				},
			},
		},
	}
	schema, err := s.Add(context.Background(), in)
	if err != nil {
		panic(err)
	}
	return schema
}
func TestService_Get(t *testing.T) {
	s := setUp[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate]("token_test")
	schema := makePerson(s)
	ret, err := s.Get(context.Background(), schema.ID)
	if err != nil {
		t.Error(err)
	}
	if schema.Name != ret.Name {
		t.Error("expected", schema.Name, "got", ret.Name)
	}
}

func TestService_List(t *testing.T) {
	s := setUp[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate]("token_test")
	for i := 0; i < 10; i++ {
		makePerson(s)
	}

	ret, err := s.List(context.Background(), &petri.TokenFilter{
		Name: &petri.Selector[string]{
			Equals: "person",
		},
	})
	if err != nil {
		t.Error(err)
	}
	if len(ret) != 10 {
		t.Error("expected", 10, "got", len(ret))
	}
}

func TestService_Update(t *testing.T) {
	s := setUp[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate]("token_test")
	schema := makePerson(s)

	_, err := s.Update(context.Background(), schema.ID, &petri.TokenUpdate{
		Name: "person2",
		Mask: petri.TokenMask{
			Name: true,
		},
	})
	if err != nil {
		t.Error(err)
	}
	updated, err := s.Get(context.Background(), schema.ID)
	if err != nil {
		t.Error(err)
	}
	if updated.Name != "person2" {
		t.Error("expected", "person2", "got", updated.Name)
	}
}

func TestService_Remove(t *testing.T) {
	s := setUp[*petri.TokenSchema, *petri.TokenInput, *petri.TokenFilter, *petri.TokenUpdate]("token_test")
	schema := makePerson(s)

	d, err := s.Remove(schema.ID)
	if err != nil {
		t.Error(err)
	}
	if d.Name != schema.Name {
		t.Error("expected", schema.Name, "got", d.Name)
	}
	if d.ID != schema.ID {
		t.Error("expected", schema.ID, "got", d.ID)
	}
}
