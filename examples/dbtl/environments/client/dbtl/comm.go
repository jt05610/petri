package main

import (
	"client/builder/proto/v1/builder"
	"client/designer/proto/v1/designer"
	"client/learner/proto/v1/learner"
	"client/tester/proto/v1/tester"
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/caser"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"time"
)

type device struct {
	designer designer.DesignerServiceClient
	builder  builder.BuilderServiceClient
	tester   tester.TesterServiceClient
	learner  learner.LearnerServiceClient
	timeout  time.Duration
}

type Method[T, U any] struct {
	Name   string
	Handle func(ctx context.Context, i T) (U, error)
}

func listMethods(name string, i interface{}) []Method[proto.Message, proto.Message] {
	var methods []Method[proto.Message, proto.Message]
	iType := reflect.TypeOf(i)
	for j := 0; j < iType.NumMethod(); j++ {
		m := iType.Method(j)
		h := Method[proto.Message, proto.Message]{
			Name: name + "." + caser.New(m.Name).CamelCase(),
			Handle: func(ctx context.Context, t proto.Message) (proto.Message, error) {
				res := m.Func.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(t),
				})
				err := res[1].Interface()
				if err != nil {
					return nil, err.(error)
				}
				return res[0].Interface().(proto.Message), nil
			},
		}
		methods = append(methods, h)
	}
	return methods
}

type handler[T, U proto.Message] struct {
	handle    func(ctx context.Context, t T) (U, error)
	timeout   time.Duration
	OutSchema *petri.TokenSchema
}

type GRPCValue[T proto.Message] struct {
	Message T
}

func (G GRPCValue[T]) Bytes() []byte {
	ret, err := proto.Marshal(G.Message)
	if err != nil {
		panic(err)
	}
	return ret
}

func (G GRPCValue[T]) FromBytes(bytes []byte) (petri.Value, error) {
	var t T
	err := proto.Unmarshal(bytes, t)
	if err != nil {
		return nil, err
	}
	G.Message = t
	return G, nil
}

func TokensToProto[T proto.Message](t map[string]petri.Token, ret T) error {
	fieldNameMap := make(map[string]protoreflect.FieldDescriptor)
	ret.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		fieldNameMap[string(fd.Name())] = fd
		return true
	})
	for field, tok := range t {
		if n, ok := fieldNameMap[field]; ok {
			ret.ProtoReflect().Set(n, protoreflect.ValueOf(tok.Value))
		}
	}
	if !ret.ProtoReflect().IsValid() {
		return fmt.Errorf("no valid field found")
	}
	return nil
}

var _ petri.Value = (*GRPCValue)(nil)

func (h *handler[T, U]) Handle(token ...petri.Token) ([]petri.Token, error) {
	ctx, can := context.WithTimeout(context.Background(), h.timeout)
	defer can()
	res, err := h.handle(ctx, token[0].Value.(T))
	if err != nil {
		return nil, err
	}
	v := GRPCValue[U]{}
	bb, err := proto.Marshal(res)
	if err != nil {
		return nil, err
	}
	val, err := v.FromBytes(bb)
	tt := make([]petri.Token, 0)
	res.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		tt = append(tt, petri.Token{
			Value: val,
		})
		return true
	})
	return tt, nil
}

type handlerFunc[T, U proto.Message] func(ctx context.Context, t T) (U, error)

func NewHandler[T, U proto.Message](d *device, f handlerFunc[T, U]) petri.Handler {
	return &handler[T, U]{
		handle:  f,
		timeout: d.timeout,
	}
}

type RPCMap map[string]handlerFunc[proto.Message, proto.Message]

func (d *device) registerHandler(n *petri.Net) error {
	mm := listMethods("designer", d.designer)
	mm = append(mm, listMethods("builder", d.builder)...)
	mm = append(mm, listMethods("tester", d.tester)...)
	mm = append(mm, listMethods("learner", d.learner)...)
	methodMap := make(RPCMap)
	for _, m := range mm {
		methodMap[m.Name] = m.Handle
	}
	for tn, t := range n.Transitions {
		hndl := methodMap[tn]
		if hndl == nil {
			continue
		}
		fmt.Printf("Registering handeler for: %s\n", tn)
		t.Handler = NewHandler(d, hndl)
	}
	return nil
}
