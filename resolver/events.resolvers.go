package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.42

import (
	"context"
	"fmt"

	"github.com/jt05610/petri"
	"github.com/jt05610/petri/graph/generated"
)

// InputSchema is the resolver for the inputSchema field.
func (r *eventSchemaInputResolver) InputSchema(ctx context.Context, obj *petri.EventInput, data string) error {
	panic(fmt.Errorf("not implemented: InputSchema - inputSchema"))
}

// OutputSchema is the resolver for the outputSchema field.
func (r *eventSchemaInputResolver) OutputSchema(ctx context.Context, obj *petri.EventInput, data string) error {
	panic(fmt.Errorf("not implemented: OutputSchema - outputSchema"))
}

// InputSchema is the resolver for the inputSchema field.
func (r *eventUpdateInputResolver) InputSchema(ctx context.Context, obj *petri.EventUpdate, data string) error {
	panic(fmt.Errorf("not implemented: InputSchema - inputSchema"))
}

// OutputSchema is the resolver for the outputSchema field.
func (r *eventUpdateInputResolver) OutputSchema(ctx context.Context, obj *petri.EventUpdate, data string) error {
	panic(fmt.Errorf("not implemented: OutputSchema - outputSchema"))
}

// EventSchemaInput returns generated.EventSchemaInputResolver implementation.
func (r *Resolver) EventSchemaInput() generated.EventSchemaInputResolver {
	return &eventSchemaInputResolver{r}
}

// EventUpdateInput returns generated.EventUpdateInputResolver implementation.
func (r *Resolver) EventUpdateInput() generated.EventUpdateInputResolver {
	return &eventUpdateInputResolver{r}
}

type eventSchemaInputResolver struct{ *Resolver }
type eventUpdateInputResolver struct{ *Resolver }