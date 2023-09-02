package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jt05610/petri/cmd/petrid/graph/generated"
	"github.com/jt05610/petri/cmd/petrid/graph/model"
	"github.com/jt05610/petri/prisma/db"
)

// StartSession is the resolver for the startSession field.
func (r *mutationResolver) StartSession(ctx context.Context, input model.StartSessionInput) (*model.Event, error) {
	session, err := r.SessionClient.Load(ctx, input.SessionID)
	if err != nil {
		return nil, err
	}
	sequence, err := r.RunClient.Load(ctx, session.RunID)
	if err != nil {
		return nil, err
	}
	r.Sequence = sequence
	r.Sequence.ExtractParameters()
	fmt.Println("parameters", input.Parameters)
	err = r.Sequence.ApplyParameters(input.Parameters)
	if err != nil {
		return nil, err
	}
	err = r.DevicesReady()
	if err != nil {
		return nil, err
	}
	return r.start(input.SessionID), nil
}

// NewSession is the resolver for the newSession field.
func (r *mutationResolver) NewSession(ctx context.Context, input model.NewSessionInput) (*model.Session, error) {
	run, err := r.RunClient.Load(ctx, input.SequenceID)
	if err != nil {
		return nil, err
	}
	devices := run.Devices()
	r.Sequence = run
	net, err := r.NetClient.Load(ctx, run.NetID)
	if err != nil {
		return nil, err
	}
	err = r.Sequence.ApplyNet(net)
	if err != nil {
		return nil, err
	}
	r.Sequence.ExtractParameters()
	if len(input.Instances) != len(devices) {
		return nil, errors.New("wrong number of instances")
	}
	instanceIDs := make([]string, len(input.Instances))
	for i, inst := range input.Instances {
		if _, ok := r.Known[inst.DeviceID]; !ok {
			return nil, errors.New("unknown device")
		}
		r.Routes[inst.DeviceID] = r.Known[inst.DeviceID][inst.InstanceID]
		instanceIDs[i] = inst.InstanceID
	}

	s, err := r.CreateSession(ctx, input.SequenceID, input.UserID, instanceIDs)
	if err != nil {
		return nil, err
	}
	return &model.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		RunID:     s.RunID,
		Active:    true,
		CreatedAt: s.CreatedAt.String(),
		UpdatedAt: s.UpdatedAt.String(),
	}, nil
}

// StopSession is the resolver for the stopSession field.
func (r *mutationResolver) StopSession(ctx context.Context, sessionID string) (*model.Session, error) {
	s, err := r.SessionClient.StopSession(ctx, sessionID, time.Now())
	if err != nil {
		return nil, err
	}
	events, err := r.eventHistory(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &model.Session{
		ID:     s.ID,
		UserID: s.UserID,
		RunID:  s.RunID,
		Active: false,
		Events: events,
	}, nil
}

// PauseSession is the resolver for the pauseSession field.
func (r *mutationResolver) PauseSession(ctx context.Context, sessionID string) (*model.Session, error) {
	s, err := r.SessionClient.PauseSession(ctx, sessionID, time.Now())
	if err != nil {
		return nil, err
	}
	return &model.Session{
		ID:     s.ID,
		UserID: s.UserID,
		RunID:  s.RunID,
		Active: false,
	}, nil
}

// ResumeSession is the resolver for the resumeSession field.
func (r *mutationResolver) ResumeSession(ctx context.Context, sessionID string) (*model.Session, error) {
	s, err := r.SessionClient.ResumeSession(ctx, sessionID, time.Now())
	if err != nil {
		return nil, err
	}
	return &model.Session{
		ID:     s.ID,
		UserID: s.UserID,
		RunID:  s.RunID,
		Active: true,
	}, nil
}

// ActiveSessions is the resolver for the activeSessions field.
func (r *queryResolver) ActiveSessions(ctx context.Context) ([]*model.Session, error) {
	s, err := r.SessionClient.ActiveSessions(ctx)
	if err != nil {
		return nil, err
	}
	sessions := make([]*model.Session, len(s))
	for i, session := range s {
		sessions[i] = &model.Session{
			ID:        session.ID,
			Active:    true,
			CreatedAt: session.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: session.UpdatedAt.Format(time.RFC3339Nano),
		}
	}
	return sessions, nil
}

// Sessions is the resolver for the sessions field.
func (r *queryResolver) Sessions(ctx context.Context, runID string) ([]*model.Session, error) {
	s, err := r.ListSessions(ctx, runID)
	if err != nil {
		return nil, err
	}
	sessions := make([]*model.Session, len(s))
	for i, session := range s {
		sessions[i] = &model.Session{
			ID:        session.ID,
			Active:    session.State == db.SessionStateRUNNING,
			CreatedAt: session.CreatedAt.String(),
			UpdatedAt: session.UpdatedAt.String(),
		}
	}
	return sessions, nil
}

// CurrentStep is the resolver for the currentStep field.
func (r *queryResolver) CurrentStep(ctx context.Context, sessionID string) (int, error) {
	return r.Controller.CurrentStep, nil
}

// EventHistory is the resolver for the eventHistory field.
func (r *queryResolver) EventHistory(ctx context.Context, sessionID string) ([]*model.Event, error) {
	return r.eventHistory(ctx, sessionID)
}

// Instances is the resolver for the instances field.
func (r *queryResolver) Instances(ctx context.Context, runID string) ([]*model.Instance, error) {
	if r.Sequence == nil {
		return nil, fmt.Errorf("no sequence")
	}
	if r.Sequence.ID != runID {
		return nil, fmt.Errorf("runID mismatch")
	}

	dd := r.Sequence.Devices()
	instances := make([]*model.Instance, 0, len(dd))
	for _, d := range dd {
		for _, instance := range r.Known[d.ID] {
			instances = append(instances, &model.Instance{
				ID:   instance.ID,
				Name: instance.ID,
			})
		}
	}
	return instances, nil
}

// Devices is the resolver for the devices field.
func (r *queryResolver) Devices(ctx context.Context, filter *string) ([]*model.Device, error) {
	ret := make([]*model.Device, 0)
	for deviceID, instances := range r.Known {
		devInstances := make([]*model.Instance, len(instances))
		i := 0
		for _, instance := range instances {
			devInstances[i] = &model.Instance{
				ID:   instance.ID,
				Name: instance.ID,
			}
			i++
		}

		ret = append(ret, &model.Device{
			ID:        deviceID,
			Name:      deviceID,
			Instances: devInstances,
		})
	}
	return ret, nil
}

// DeviceMarkings is the resolver for the deviceMarkings field.
func (r *queryResolver) DeviceMarkings(ctx context.Context, input model.DeviceMarkingsInput) ([]*model.DeviceMarking, error) {
	ret := make([]*model.DeviceMarking, 0)
	for _, device := range input.Instances {
		ret = append(ret, &model.DeviceMarking{
			DeviceID: device.DeviceID,
			Marking:  r.Routes[device.DeviceID].Marking.JSON(),
		})
	}
	fmt.Printf("DeviceMarkings: %v\n", ret)
	return ret, nil
}

// NewEvents is the resolver for the newEvents field.
func (r *queryResolver) NewEvents(ctx context.Context, sessionID string) ([]*model.Event, error) {
	return r.newEvents(sessionID)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
