package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.42

import (
	"context"
	"fmt"

	"github.com/jt05610/petri"
	"github.com/jt05610/petri/graph/generated"
	"github.com/jt05610/petri/resolver/model"
)

// Name is the resolver for the name field.
func (r *arcResolver) Name(ctx context.Context, obj *petri.Arc) (string, error) {
	panic(fmt.Errorf("not implemented: Name - name"))
}

// Properties is the resolver for the properties field.
func (r *arcResolver) Properties(ctx context.Context, obj *petri.Arc) (model.JSON, error) {
	panic(fmt.Errorf("not implemented: Properties - properties"))
}

// Start is the resolver for the start field.
func (r *arcFilterInputResolver) Start(ctx context.Context, obj *petri.ArcFilter, data *string) error {
	panic(fmt.Errorf("not implemented: Start - start"))
}

// End is the resolver for the end field.
func (r *arcFilterInputResolver) End(ctx context.Context, obj *petri.ArcFilter, data *string) error {
	panic(fmt.Errorf("not implemented: End - end"))
}

// Name is the resolver for the name field.
func (r *arcInputResolver) Name(ctx context.Context, obj *petri.ArcInput, data string) error {
	panic(fmt.Errorf("not implemented: Name - name"))
}

// Properties is the resolver for the properties field.
func (r *arcInputResolver) Properties(ctx context.Context, obj *petri.ArcInput, data model.JSON) error {
	panic(fmt.Errorf("not implemented: Properties - properties"))
}

// Name is the resolver for the name field.
func (r *arcUpdateInputResolver) Name(ctx context.Context, obj *petri.ArcUpdate, data *string) error {
	panic(fmt.Errorf("not implemented: Name - name"))
}

// Type is the resolver for the type field.
func (r *arcUpdateInputResolver) Type(ctx context.Context, obj *petri.ArcUpdate, data *string) error {
	panic(fmt.Errorf("not implemented: Type - type"))
}

// Properties is the resolver for the properties field.
func (r *arcUpdateInputResolver) Properties(ctx context.Context, obj *petri.ArcUpdate, data model.JSON) error {
	panic(fmt.Errorf("not implemented: Properties - properties"))
}

// Arc returns generated.ArcResolver implementation.
func (r *Resolver) Arc() generated.ArcResolver { return &arcResolver{r} }

// ArcFilterInput returns generated.ArcFilterInputResolver implementation.
func (r *Resolver) ArcFilterInput() generated.ArcFilterInputResolver {
	return &arcFilterInputResolver{r}
}

// ArcInput returns generated.ArcInputResolver implementation.
func (r *Resolver) ArcInput() generated.ArcInputResolver { return &arcInputResolver{r} }

// ArcUpdateInput returns generated.ArcUpdateInputResolver implementation.
func (r *Resolver) ArcUpdateInput() generated.ArcUpdateInputResolver {
	return &arcUpdateInputResolver{r}
}

type arcResolver struct{ *Resolver }
type arcFilterInputResolver struct{ *Resolver }
type arcInputResolver struct{ *Resolver }
type arcUpdateInputResolver struct{ *Resolver }