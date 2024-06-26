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

// Properties is the resolver for the properties field.
func (r *netResolver) Properties(ctx context.Context, obj *petri.Net) (model.JSON, error) {
	panic(fmt.Errorf("not implemented: Properties - properties"))
}

// Properties is the resolver for the properties field.
func (r *netInputResolver) Properties(ctx context.Context, obj *petri.NetInput, data model.JSON) error {
	panic(fmt.Errorf("not implemented: Properties - properties"))
}

// Name is the resolver for the name field.
func (r *netUpdateInputResolver) Name(ctx context.Context, obj *petri.NetUpdate, data *string) error {
	panic(fmt.Errorf("not implemented: Name - name"))
}

// Type is the resolver for the type field.
func (r *netUpdateInputResolver) Type(ctx context.Context, obj *petri.NetUpdate, data *string) error {
	panic(fmt.Errorf("not implemented: Type - type"))
}

// Properties is the resolver for the properties field.
func (r *netUpdateInputResolver) Properties(ctx context.Context, obj *petri.NetUpdate, data model.JSON) error {
	panic(fmt.Errorf("not implemented: Properties - properties"))
}

// Net returns generated.NetResolver implementation.
func (r *Resolver) Net() generated.NetResolver { return &netResolver{r} }

// NetInput returns generated.NetInputResolver implementation.
func (r *Resolver) NetInput() generated.NetInputResolver { return &netInputResolver{r} }

// NetUpdateInput returns generated.NetUpdateInputResolver implementation.
func (r *Resolver) NetUpdateInput() generated.NetUpdateInputResolver {
	return &netUpdateInputResolver{r}
}

type netResolver struct{ *Resolver }
type netInputResolver struct{ *Resolver }
type netUpdateInputResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *netResolver) Children(ctx context.Context, obj *petri.Net) ([]*petri.Net, error) {
	panic(fmt.Errorf("not implemented: Children - children"))
}
