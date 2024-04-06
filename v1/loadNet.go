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
	"net"
	"os"
	"path/filepath"
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

type registerFunc[T any] func(registrar grpc.ServiceRegistrar, srv T)

func Serve[T any](ctx context.Context, host string, n *petri.Net, srv T, reg registerFunc[T]) error {
	ctx, can := context.WithCancel(context.Background())
	defer can()
	server := grpc.NewServer()
	reg(server, srv)
	lis, err := net.Listen("tcp", host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("{\"%s\": \"%s\"}\n", n.Name, lis.Addr().String())
	return server.Serve(lis)
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
