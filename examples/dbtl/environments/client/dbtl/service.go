package main

import (
	"client/dbtl/proto/v1/dbtl"
	"context"
	"github.com/jt05610/petri"
	"google.golang.org/protobuf/proto"
)

type MarkingChange struct {
	Place  string
	Tokens []petri.Token
}

type service struct {
	dbtl.UnimplementedDbtlServiceServer
	*petri.Net
	petri.Marking
	markingRequest chan MarkingChange
}

func (s *service) update() error {
	var err error
	s.Marking, err = s.Process(s.Marking)
	if err != nil {
		return err
	}
	return nil
}

func enqueue[T proto.Message](s *service, ctx context.Context, request T, place string) error {
	pl := s.Place(place)
	bb, err := proto.Marshal(request)
	if err != nil {
		return err
	}
	tok, err := pl.AcceptedTokens[0].NewToken(bb)
	if err != nil {
		return err
	}
	err = s.Marking[place].Enqueue(ctx, tok)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) LoadDesignSpace(ctx context.Context, request *dbtl.LoadDesignSpaceRequest) (*dbtl.LoadDesignSpaceResponse, error) {
	err := enqueue(s, ctx, request, "designer.loadDesignSpaceInput")
	if err != nil {
		return nil, err
	}
	err = s.update()
	if err != nil {
		return nil, err
	}
	return &dbtl.LoadDesignSpaceResponse{}, nil
}

func (s *service) LoadSample(ctx context.Context, request *dbtl.LoadSampleRequest) (*dbtl.LoadSampleResponse, error) {
	err := enqueue(s, ctx, request, "designer.loadSampleInput")
	if err != nil {
		return nil, err
	}
	err = s.update()
	if err != nil {
		return nil, err
	}
	return &dbtl.LoadSampleResponse{}, nil
}

func (s *service) NewDesign(ctx context.Context, request *dbtl.NewDesignRequest) (*dbtl.NewDesignResponse, error) {
	err := enqueue(s, ctx, request, "designer.newDesignInput")
	if err != nil {
		return nil, err
	}
	err = s.update()
	if err != nil {
		return nil, err
	}
	return &dbtl.NewDesignResponse{}, nil
}

func (s *service) Recipe(ctx context.Context, request *dbtl.RecipeRequest) (*dbtl.RecipeResponse, error) {
	err := enqueue(s, ctx, request, "tester.recipeInput")
	if err != nil {
		return nil, err
	}
	err = s.update()
	if err != nil {
		return nil, err
	}
	return &dbtl.RecipeResponse{}, nil
}

var _ dbtl.DbtlServiceServer = (*service)(nil)
