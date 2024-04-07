package petri

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/builder"
	"github.com/jt05610/petri/petrifile/v1/yaml"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
	"os"
	"path/filepath"
	"sync"
)

func LoadNet(fName string) *petri.Net {
	parDir := filepath.Dir(fName)
	bld := builder.NewBuilder(nil, ".", parDir)
	r := yaml.NewService(bld)
	bld = bld.WithService("yaml", r)
	bld = bld.WithService("yml", r)
	f, err := os.Open(fName)
	if err != nil {
		panic(err)
	}

	n, err := r.Load(context.Background(), f)
	if err != nil {
		panic(err)
	}
	return n
}

type ProtobufService struct {
	msgType protoreflect.MessageType
}

func NewProtobufService(schema *petri.TokenSchema, msgType protoreflect.MessageType) *ProtobufService {
	return &ProtobufService{
		msgType: msgType,
	}
}

func (p *ProtobufService) Load(_ context.Context, rdr io.Reader) (petri.Token, error) {
	var tok petri.Token
	bb := make([]byte, 1024)
	t := p.msgType.New().Interface()
	n, err := rdr.Read(bb)
	if err != nil {
		return tok, err
	}
	err = proto.Unmarshal(bb[:n], t)
	if err != nil {
		return tok, err
	}
	buf := bytes.NewBuffer(bb[:n])
	tok = petri.Token{
		Value: buf,
	}
	return tok, nil
}

func (p *ProtobufService) Flush(_ context.Context, wr io.Writer, token petri.Token) error {
	_, err := io.Copy(wr, token.Value)
	return err
}

var _ petri.TokenService = (*ProtobufService)(nil)

func RegisterProtobufService(n *petri.Net, schema string, msgType protoreflect.MessageType) error {
	s, found := n.TokenSchemas[schema]
	if !found {
		return fmt.Errorf("schema %s not found", schema)
	}
	return n.Register(schema, NewProtobufService(s, msgType))
}

type TokenTypeMap map[*petri.TokenSchema]protoreflect.MessageType

func RegisterProtobufServices(n *petri.Net, types TokenTypeMap) error {
	for schema, msgType := range types {
		err := RegisterProtobufService(n, schema.Name, msgType)
		if err != nil {
			return err
		}
	}
	return nil
}

type TransitionHandler[T, U proto.Message] func(ctx context.Context, in T) (U, error)

type OptionService struct {
	opt []grpc.CallOption
	mu  sync.RWMutex
}

func (s *OptionService) Set(opts ...grpc.CallOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.opt = opts
}

func (s *OptionService) Get() []grpc.CallOption {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.opt
}

func (s *OptionService) Append(opts ...grpc.CallOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.opt = append(s.opt, opts...)
}

func (s *OptionService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.opt = nil
}

func NewOptionService() *OptionService {
	return &OptionService{
		opt: make([]grpc.CallOption, 0),
	}
}

func TransitionClient[T, U proto.Message](f func(ctx context.Context, in T, opts ...grpc.CallOption) (U, error), srv *OptionService) TransitionHandler[T, U] {
	return func(ctx context.Context, in T) (U, error) {
		return f(ctx, in, srv.Get()...)
	}
}

type TokenConverter[T proto.Message] struct {
	msgType protoreflect.MessageType
	fields  map[string]protoreflect.FieldDescriptor
}

func (t *TokenConverter[T]) Marshal(msg T) (petri.TokenMap, error) {
	tokenMap := make(petri.TokenMap)
	var err error
	msg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		bb, err := proto.Marshal(v.Message().Interface())
		if err != nil {
			return false
		}
		tokenMap[string(fd.Name())] = petri.Token{
			Value: bytes.NewBuffer(bb),
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return tokenMap, nil
}

func (t *TokenConverter[T]) Unmarshal(tok petri.TokenMap) (T, error) {
	msg := t.msgType.New()
	for k, v := range tok {
		field, found := t.fields[k]
		if !found {
			continue
		}
		bb, err := io.ReadAll(v.Value)
		if err != nil {
			var zero T
			return zero, err
		}
		fieldMsg := msg.Mutable(field)
		err = proto.Unmarshal(bb, fieldMsg.Message().Interface())
		if err != nil {
			var zero T
			return zero, err
		}

	}
	return msg.Interface().(T), nil
}

func (t *TokenConverter[T]) discoverFields() {
	t.fields = make(map[string]protoreflect.FieldDescriptor)
	fields := t.msgType.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		t.fields[string(f.Name())] = f
	}
}

func NewTokenConverter[T proto.Message]() *TokenConverter[T] {
	var t T
	tc := &TokenConverter[T]{
		msgType: t.ProtoReflect().Type(),
	}
	tc.discoverFields()
	return tc
}

func (t *TokenConverter[T]) Fields() []protoreflect.FieldDescriptor {
	fields := make([]protoreflect.FieldDescriptor, len(t.fields))
	for _, field := range t.fields {
		fields[field.Number()-1] = field
	}
	return fields
}

func (t TransitionHandler[T, U]) Wrap(inCvt *TokenConverter[T], outCvt *TokenConverter[U]) petri.HandlerFunc {
	return func(ctx context.Context, in petri.TokenMap) (petri.TokenMap, error) {
		inMsg, err := inCvt.Unmarshal(in)
		if err != nil {
			return nil, err
		}
		outMsg, err := t(ctx, inMsg)
		if err != nil {
			return nil, err
		}
		return outCvt.Marshal(outMsg)
	}
}

type Handler[T, U proto.Message] struct {
	TransitionHandler[T, U]
	InCvt  *TokenConverter[T]
	OutCvt *TokenConverter[U]
	handle petri.HandlerFunc
}

func NewHandler[T, U proto.Message](h TransitionHandler[T, U]) *Handler[T, U] {
	inCvt := NewTokenConverter[T]()
	outCvt := NewTokenConverter[U]()
	hndl := &Handler[T, U]{
		TransitionHandler: h,
		InCvt:             inCvt,
		OutCvt:            outCvt,
	}
	hndl.handle = hndl.Wrap()
	return hndl
}

func (h *Handler[T, U]) Wrap() petri.HandlerFunc {
	return h.TransitionHandler.Wrap(h.InCvt, h.OutCvt)
}

func (h *Handler[T, U]) Handle(ctx context.Context, in petri.TokenMap) (petri.TokenMap, error) {
	ret, err := h.handle(ctx, in)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func RegisterTransitionHandler[T, U proto.Message](n *petri.Net, t string, handler *Handler[T, U]) error {
	transition, found := n.Transitions[t]
	if !found {
		return fmt.Errorf("transition %s not found", t)
	}
	handler.handle = handler.Wrap()
	transition.Handler = handler
	return nil
}
