package petri

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	proto "github.com/jt05610/petri/proto/v1"
	"google.golang.org/grpc"
	"io"
	"log/slog"
	"net"
)

var _ proto.PetriNetServer = (*Server)(nil)

func MergeMaps[T comparable, U any](maps ...map[T]U) map[T]U {
	ret := make(map[T]U)
	for _, m := range maps {
		for k, v := range m {
			ret[k] = v
		}
	}
	return ret
}

type Server struct {
	TokenService
	proto.UnimplementedPetriNetServer
}

func ToToken(token *proto.Token) petri.Token {
	return petri.Token{
		ID:    token.Id,
		Value: bytes.NewBuffer(token.Data),
	}
}

func (s *Server) PutToken(ctx context.Context, request *proto.PutTokenRequest) (*proto.PutTokenResponse, error) {
	err := s.TokenService.PutToken(ctx, request.PlaceId, ToToken(request.Token))
	if err != nil {
		return nil, err
	}
	return &proto.PutTokenResponse{
		Success: true,
	}, nil
}

func (s *Server) PopToken(ctx context.Context, request *proto.PopTokenRequest) (*proto.PopTokenResponse, error) {
	token, err := s.TokenService.PopToken(ctx, request.PlaceId)
	if err != nil {
		return nil, err
	}
	if request.WithValue == nil {
		return &proto.PopTokenResponse{
			Token: &proto.Token{
				Id: token.ID,
			},
		}, nil
	}
	if *request.WithValue {
		bb, err := io.ReadAll(token.Value)
		if err != nil {
			return nil, err
		}
		return &proto.PopTokenResponse{
			Token: &proto.Token{
				Id:   token.ID,
				Data: bb,
			},
		}, nil
	}
	return &proto.PopTokenResponse{
		Token: &proto.Token{
			Id: token.ID,
		},
	}, nil
}

func ConvertMarking(m petri.Marking) (*proto.Marking, error) {
	placeMarkings := make([]*proto.PlaceMarking, 0, len(m))
	for k, v := range m {
		placeMarking := &proto.PlaceMarking{
			PlaceId: k,
			Tokens:  make([]*proto.Token, 0, len(v)),
		}
		for _, t := range v {
			tok := &proto.Token{
				Id: t.ID,
			}
			if t.Value == nil {
				placeMarking.Tokens = append(placeMarking.Tokens, tok)
				continue
			}
			bb, err := io.ReadAll(t.Value)
			if err != nil {
				return nil, err
			}
			tok.Data = bb
			placeMarking.Tokens = append(placeMarking.Tokens, tok)
		}
		placeMarkings = append(placeMarkings, placeMarking)
	}
	return &proto.Marking{
		PlaceMarkings: placeMarkings,
	}, nil
}

func (s *Server) GetMarking(ctx context.Context, request *proto.GetMarkingRequest) (*proto.GetMarkingResponse, error) {
	if request.WithValue == nil {
		request.WithValue = new(bool)
		*request.WithValue = false
	}
	marking, err := s.TokenService.GetMarking(ctx, *request.WithValue)
	if err != nil {
		return nil, err
	}
	protoMarking, err := ConvertMarking(marking)
	if err != nil {
		return nil, err
	}
	return &proto.GetMarkingResponse{
		Marking: protoMarking,
	}, nil
}

func (s *Server) GetToken(ctx context.Context, request *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	token, err := s.TokenService.GetToken(ctx, request.TokenId)
	if err != nil {
		return nil, err
	}
	bb, err := io.ReadAll(token.Value)
	if err != nil {
		return nil, err
	}
	return &proto.GetTokenResponse{
		Token: &proto.Token{
			Id:   token.ID,
			Data: bb,
		},
	}, nil
}

type registerFunc[T any] func(registrar grpc.ServiceRegistrar, srv T)

type HandlerMap map[string]petri.Handler

type ServerOptions[T any] struct {
	Host         string
	Server       T
	Reg          registerFunc[T]
	Net          *petri.Net
	RPCOpts      []grpc.ServerOption
	TokenStorage TokenService
	Handlers     HandlerMap
}

type RPCServer struct {
	net.Listener
	*Server
	*petri.Net
	srv *grpc.Server
}

func (s *RPCServer) Close() {
	s.srv.Stop()
}

func (s *RPCServer) Serve(ctx context.Context) error {
	updates := make(chan RawMarking)
	ch, err := s.Monitor(ctx, updates)
	if err != nil {
		return err
	}
	go func() {
		defer close(updates)
		for {
			select {
			case <-ctx.Done():
				return
			case upd := <-ch:
				m, err := s.Net.Process(ctx, upd.Marking())
				if err != nil {
					if errors.Is(err, petri.ErrNoEvents) {
						continue
					}
					slog.Error("error processing marking", slog.Any("message", err))
				}
				updates <- MakeRawMarking(m)
			}
		}
	}()
	return s.srv.Serve(s.Listener)
}

func NewServer[T any](opt *ServerOptions[T]) (*RPCServer, error) {
	if opt.TokenStorage == nil {
		opt.TokenStorage = DefaultTokenService()
	}
	for name, handler := range opt.Handlers {
		transition := opt.Net.Transition(name)
		if transition == nil {
			return nil, fmt.Errorf("transition %s not found", name)
		}
		transition.Handler = handler
	}
	server := grpc.NewServer(opt.RPCOpts...)
	opt.Reg(server, opt.Server)
	s := &Server{TokenService: opt.TokenStorage}
	proto.RegisterPetriNetServer(server, s)
	lis, err := net.Listen("tcp", opt.Host)
	if err != nil {
		panic(err)
	}
	return &RPCServer{
		Server:   s,
		Net:      opt.Net,
		srv:      server,
		Listener: lis,
	}, nil
}

type RPCClient proto.PetriNetClient

type ClientOptions[T RPCClient] struct {
	Addr       string
	Attach     func(grpc.ClientConnInterface) (T, error)
	RPCOptions []grpc.DialOption
}

func Dial[T RPCClient](ctx context.Context, opt *ClientOptions[T]) (T, error) {
	conn, err := grpc.DialContext(ctx, opt.Addr, opt.RPCOptions...)
	if err != nil {
		panic(err)
	}
	ret, err := opt.Attach(conn)
	if err != nil {
		var zero T
		return zero, err
	}
	return ret, nil
}

func NewClient(conn grpc.ClientConnInterface) proto.PetriNetClient {
	return proto.NewPetriNetClient(conn)
}
