package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"
	"fmt"

	"github.com/jt05610/petri/cmd/petrid/graph/generated"
	"github.com/jt05610/petri/cmd/petrid/graph/model"
)

// NewSession is the resolver for the newSession field.
func (r *mutationResolver) NewSession(ctx context.Context, input model.NewSessionInput) (*model.Session, error) {
	panic(fmt.Errorf("not implemented: NewSession - newSession"))
}

// StartSession is the resolver for the startSession field.
func (r *mutationResolver) StartSession(ctx context.Context, input model.StartSessionInput) (*model.Session, error) {
	panic(fmt.Errorf("not implemented: StartSession - startSession"))
}

// StopSession is the resolver for the stopSession field.
func (r *mutationResolver) StopSession(ctx context.Context, sessionID string) (*model.Session, error) {
	panic(fmt.Errorf("not implemented: StopSession - stopSession"))
}

// DeleteSession is the resolver for the deleteSession field.
func (r *mutationResolver) DeleteSession(ctx context.Context, sessionID string) (*model.HandleResult, error) {
	panic(fmt.Errorf("not implemented: DeleteSession - deleteSession"))
}

// ActiveSessions is the resolver for the activeSessions field.
func (r *queryResolver) ActiveSessions(ctx context.Context) ([]*model.Session, error) {
	panic(fmt.Errorf("not implemented: ActiveSessions - activeSessions"))
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, sessionID string) (*model.Session, error) {
	panic(fmt.Errorf("not implemented: Session - session"))
}

// Sessions is the resolver for the sessions field.
func (r *queryResolver) Sessions(ctx context.Context, runID string) ([]*model.Session, error) {
	s, err := r.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	sessions := make([]*model.Session, len(s))
	for i, session := range s {
		sessions[i] = &model.Session{
			ID:        session.ID,
			Active:    false,
			CreatedAt: session.CreatedAt.String(),
			UpdatedAt: session.UpdatedAt.String(),
		}
	}
	return sessions, nil
}

// CurrentStep is the resolver for the currentStep field.
func (r *queryResolver) CurrentStep(ctx context.Context, sessionID string) (int, error) {
	panic(fmt.Errorf("not implemented: CurrentStep - currentStep"))
}

// EventHistory is the resolver for the eventHistory field.
func (r *queryResolver) EventHistory(ctx context.Context, sessionID string) ([]*model.Event, error) {
	panic(fmt.Errorf("not implemented: EventHistory - eventHistory"))
}

// Instances is the resolver for the instances field.
func (r *queryResolver) Instances(ctx context.Context, runID string) ([]*model.Instance, error) {
	panic(fmt.Errorf("not implemented: Instances - instances"))
}

// Devices is the resolver for the devices field.
func (r *queryResolver) Devices(ctx context.Context, filter *string) ([]*model.Device, error) {
	panic(fmt.Errorf("not implemented: Devices - devices"))
}

// Event is the resolver for the event field.
func (r *subscriptionResolver) Event(ctx context.Context, sessionID string) (<-chan *model.Event, error) {
	panic(fmt.Errorf("not implemented: Event - event"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
