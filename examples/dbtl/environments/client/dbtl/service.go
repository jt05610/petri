package main

import (
	"client/dbtl/proto/v1/dbtl"
	"context"
	"errors"
	"github.com/jt05610/petri"
	"google.golang.org/protobuf/proto"
)

type MarkingChange struct {
	Place  string
	Tokens []petri.Token
}

type Service struct {
	dbtl.UnimplementedDbtlServiceServer
	*petri.Net
	petri.Marking
	markingRequest chan MarkingChange
	markingUpdated chan petri.Marking
}

func (s *Service) Listen(ctx context.Context) error {
	for {
		select {
		case change := <-s.markingRequest:
			mark := s.Marking
			mark[change.Place] = change.Tokens
			upd, err := s.Process(mark)
			if err != nil {
				return err
			}
			select {
			case s.markingUpdated <- upd:
			default:
			}
		case <-ctx.Done():
			if !errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}
		}
	}
}

func enqueue[T proto.Message](s *Service, ctx context.Context, request T, place string) error {
	pl := s.Place(place)
	if pl == nil {
		return errors.New("place not found")
	}
	bb, err := proto.Marshal(request)
	if err != nil {
		return err
	}
	tok, err := pl.AcceptedTokens[0].NewToken(bb)
	if err != nil {
		return err
	}
	mark := s.Marking
	tt := append(mark[place], tok)
	s.markingRequest <- MarkingChange{
		Place:  place,
		Tokens: tt,
	}
	if err != nil {
		return err
	}
	select {
	case <-s.markingUpdated:
		return nil
	case <-ctx.Done():
		return errors.New("context canceled")
	}
}

func (s *Service) LoadDesignSpace(ctx context.Context, request *dbtl.LoadDesignSpaceRequest) (*dbtl.LoadDesignSpaceResponse, error) {
	err := enqueue(s, ctx, request, "designer.loadDesignSpaceInput")
	if err != nil {
		return nil, err
	}
	return &dbtl.LoadDesignSpaceResponse{}, nil
}

func (s *Service) LoadSample(ctx context.Context, request *dbtl.LoadSampleRequest) (*dbtl.LoadSampleResponse, error) {
	err := enqueue(s, ctx, request, "designer.loadSampleInput")
	if err != nil {
		return nil, err
	}
	return &dbtl.LoadSampleResponse{}, nil
}

func (s *Service) NewDesign(ctx context.Context, request *dbtl.NewDesignRequest) (*dbtl.NewDesignResponse, error) {
	err := enqueue(s, ctx, request, "designer.newDesignInput")
	if err != nil {
		return nil, err
	}

	return &dbtl.NewDesignResponse{}, nil
}

func (s *Service) Recipe(ctx context.Context, request *dbtl.RecipeRequest) (*dbtl.RecipeResponse, error) {
	err := enqueue(s, ctx, request, "tester.recipeInput")
	if err != nil {
		return nil, err
	}
	return &dbtl.RecipeResponse{}, nil
}

var _ dbtl.DbtlServiceServer = (*Service)(nil)

func NewService(n *petri.Net, initialMarking petri.Marking) *Service {
	return &Service{
		Net:            n,
		Marking:        initialMarking,
		markingRequest: make(chan MarkingChange),
		markingUpdated: make(chan petri.Marking),
	}
}
