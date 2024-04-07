package petri

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type TokenType string

type FloatType struct {
	min *float64
	max *float64
}

func (f *FloatType) Properties() Properties {
	return Properties{
		Type: Float,
	}
}

func (f *FloatType) String() string {
	return "float"
}

func (f *FloatType) IsValid(value interface{}) bool {
	_, ok := value.(float64)
	return ok
}

type IntegerType struct{}

func (i *IntegerType) Properties() Properties {
	return Properties{
		Type: Int,
	}
}

func (i *IntegerType) String() string {
	return "int"
}

func (i *IntegerType) IsValid(value interface{}) bool {
	_, ok := value.(int64)
	return ok
}

type StringType struct{}

func (s *StringType) Properties() Properties {
	return Properties{
		Type: Str,
	}
}

func (s *StringType) String() string {
	return "string"
}

func (s *StringType) IsValid(value interface{}) bool {
	_, ok := value.(string)
	return ok
}

type BooleanType struct{}

func (b *BooleanType) Properties() Properties {
	return Properties{
		Type: Bool,
	}
}

func (b *BooleanType) String() string {
	return "boolean"
}

func (b *BooleanType) IsValid(value interface{}) bool {
	_, ok := value.(bool)
	return ok
}

type Properties struct {
	Name       string                `json:"-"`
	Type       TokenType             `json:"type"`
	Properties map[string]Properties `json:"properties,omitempty"`
}

type ObjectType struct {
	Props map[string]Properties
}

func (o *ObjectType) Properties() Properties {
	return Properties{
		Type:       Obj,
		Properties: o.Props,
	}
}

func (o *ObjectType) String() string {
	return "object"
}

var (
	Float     = TokenType("float")
	Int       = TokenType("int")
	Str       = TokenType("string")
	Bool      = TokenType("bool")
	Obj       = TokenType("object")
	Sig       = TokenType("signal")
	TimeStamp = TokenType("time")
	Arr       = TokenType("array")
)

func (t TokenType) IsPrimitive() bool {
	return t == Float || t == Int || t == Str || t == Bool || t == Sig || t == TimeStamp
}

func (t TokenType) IsValid(value interface{}) bool {
	switch t {
	case Float:
		_, ok := value.(float64)
		return ok
	case Int:
		_, ok := value.(int64)
		return ok
	case Str:
		_, ok := value.(string)
		return ok
	case Bool:
		_, ok := value.(bool)
		return ok
	case Sig:
		return value != nil
	case Obj:
		_, ok := value.(map[string]interface{})
		return ok
	}
	return false
}

type Indexable interface {
	Index() string
}

type TokenService interface {
	Load(ctx context.Context, rdr io.Reader) (Token, error)
	Flush(ctx context.Context, wr io.Writer, token Token) error
}

// TokenSchema is a simple struct that describes a token in a Petri net. Petri net operations are
// performed on tokens, and tokens are the only objects that can be placed in a Petri net.
type TokenSchema struct {
	// ID is the unique identifier of the token schema.
	ID string `json:"_id"`
	// Name is the name of the token schema.
	Name string `json:"name"`
	// Type is the type of the token schema.
	Type       TokenType             `json:"type"`
	Properties map[string]Properties `json:"properties,omitempty"`
	TokenService
	Package string
}

func NewTokenSchema(name string) *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: name,
		Type: Sig,
	}
}

func (t *TokenSchema) WithType(ty TokenType) *TokenSchema {
	t.Type = ty
	return t
}

func (t *TokenSchema) WithProperties(props map[string]Properties) *TokenSchema {
	t.Type = Obj
	t.Properties = props
	return t
}

func (t *TokenSchema) Index() string {
	return t.String()
}

func (t *TokenSchema) PostInit() error {
	return nil
}

func (t *TokenSchema) PropertiesJSON() map[string]string {
	if t.Properties != nil {
		ret := make(map[string]string)
		for key, value := range t.Properties {
			ret[key] = string(value.Type)
		}
		return ret
	}
	return nil
}

func (t *TokenSchema) Document() Document {
	if t.Properties != nil {
		return Document{
			"_id":        t.ID,
			"name":       t.Name,
			"type":       t.Type,
			"properties": t.Properties,
		}
	}
	return Document{
		"_id":  t.ID,
		"name": t.Name,
		"type": t.Type,
	}
}

func (t *TokenSchema) String() string {
	return t.Name
}

type Serializer interface {
	Bytes() []byte
}

type Deserializer interface {
	FromBytes([]byte) (Value, error)
}
type Value interface {
	Bytes() []byte
	FromBytes([]byte) (Value, error)
}

// Token is an instance of a TokenSchema.
type Token struct {
	// ID is the unique identifier of the token.
	ID string `json:"_id"`
	// Schema is the schema of the token.
	Schema *TokenSchema `json:"schema"`
	// Value is the value of the token.
	Value io.Reader `json:"value"`
}

func (t Token) String() string {
	bb := make([]byte, 1024)
	n, err := t.Value.Read(bb)
	if err != nil {
		return ""
	}
	buf := bytes.NewBuffer(bb[:n])
	t.Value = buf
	return string(bb[:n])
}

func peekBytes(r io.Reader) ([]byte, io.Reader) {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	bb, err := io.ReadAll(tee)
	if err != nil {
		return nil, r
	}
	return bb, buf
}

func (t Token) Equals(token Token) bool {
	var bb1, bb2 []byte
	bb1, t.Value = peekBytes(t.Value)
	bb2, token.Value = peekBytes(token.Value)
	if !bytes.Equal(bb1, bb2) {
		return false
	}
	return true
}

func (t *TokenSchema) Kind() Kind {
	return TokenObject
}

func (t *TokenSchema) Identifier() string {
	return t.ID
}

func parseTokenType(t string) (TokenType, error) {
	switch t {
	case "float":
		return Float, nil
	case "integer":
		return Int, nil
	case "string":
		return Str, nil
	case "boolean":
		return Bool, nil
	case "object":
		return Obj, nil
	default:
		return "", errors.New("invalid token type")
	}
}

type InvalidTokenValueError struct {
	TokenSchema *TokenSchema
	Value       interface{}
}

func (e *InvalidTokenValueError) Error() string {
	return fmt.Sprintf("invalid value for token %s: %v", e.TokenSchema.Name, e.Value)
}

// NewToken creates a new token from the schema.
func (t *TokenSchema) NewToken(rdr io.Reader) (Token, error) {
	v, err := t.TokenService.Load(context.Background(), rdr)
	if err != nil {
		return Token{}, err
	}
	buf := new(bytes.Buffer)
	err = t.Flush(context.Background(), buf, v)
	if err == nil {
		return Token{}, err
	}

	tok := Token{
		ID:     ID(),
		Schema: t,
		Value:  buf,
	}
	return tok, nil
}

type TokenMap map[string]Token

type Handler interface {
	Handle(ctx context.Context, tokens TokenMap) (TokenMap, error)
}

type HandlerFunc func(ctx context.Context, tokens TokenMap) (TokenMap, error)

type StringValue struct {
	value string
}

func (s StringValue) Bytes() []byte {
	return []byte(s.value)
}

func (s StringValue) FromBytes(bytes []byte) (Value, error) {
	s.value = string(bytes)
	return s, nil
}

var (
	_ Value = StringValue{}
	_ Value = SignalValue{}
	_ Value = IntValue{}
)

type SignalValue struct {
}

func (s SignalValue) Bytes() []byte {
	return []byte{1}
}

func (s SignalValue) FromBytes(b []byte) (Value, error) {
	if b == nil {
		return nil, errors.New("nil signal")
	}
	return s, nil
}

type IntValue struct {
	value int
}

func (i IntValue) Bytes() []byte {
	return []byte(strconv.Itoa(i.value))
}

func (i IntValue) FromBytes(b []byte) (Value, error) {
	val, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, err
	}
	i.value = val
	return i, nil
}

func (i IntValue) Value() int {
	return i.value
}

func NewIntValue(value int) IntValue {
	return IntValue{value: value}
}

func Signal() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "signal",
		Type: Sig,
	}
}

func String() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "string",
		Type: Str,
	}
}

func Float64() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "float",
		Type: Float,
	}
}

func Integer() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "int",
		Type: Int,
	}
}

func Boolean() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "bool",
		Type: Bool,
	}
}

func Time() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "time",
		Type: TimeStamp,
	}
}

func Array(t *TokenSchema) *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "array",
		Type: Obj,
		Properties: map[string]Properties{
			"items": {
				Type: t.Type,
			},
		},
	}
}

type Field struct {
	Name string
	Type TypeNode
}

type TypeNode struct {
	Name    string
	Fields  []Field
	IsArray bool
}

type TypeGraph struct {
	Nodes map[string]TypeNode
}

func (t *TypeGraph) AddNode(n TypeNode) {
	t.Nodes[n.Name] = n
}

func NodeFromType(name string, t Type) TypeNode {
	ret := TypeNode{
		Name:   name,
		Fields: make([]Field, 0, len(t)),
	}
	for k, v := range t {
		arr := false
		if IsArray(v) {
			v = ArrayOf(v)
			arr = true
		}
		ret.Fields = append(ret.Fields, Field{
			Name: k,
			Type: TypeNode{
				Name:    v,
				IsArray: arr,
			},
		})
	}
	return ret
}

type Type map[string]string

func IsArray(t string) bool {
	return strings.HasPrefix(t, "[]")
}

func ArrayOf(t string) string {
	return strings.TrimPrefix(t, "[]")
}

func (t Type) Schema(name string, lookup map[string]map[string]Properties) *TokenSchema {
	s := NewTokenSchema(name)
	props := make(map[string]Properties)
	for k, v := range t {
		if TokenType(v).IsPrimitive() {
			props[k] = Properties{
				Type:       TokenType(v),
				Properties: nil,
			}
			continue
		}
		if IsArray(v) {
			v = ArrayOf(v)
			if TokenType(v).IsPrimitive() {
				props[k] = Properties{
					Type: Arr,
					Properties: map[string]Properties{
						"items": {
							Type: TokenType(v),
						},
					},
				}
			} else {
				p, ok := lookup[v]
				if !ok {
					panic(fmt.Sprintf("unknown type %s. Please declare it", v))
				}
				props[k] = Properties{
					Type: Arr,
					Properties: map[string]Properties{
						"items": {
							Type:       Obj,
							Properties: p,
						},
					},
				}
			}
		} else {
			p, ok := lookup[v]
			if !ok {
				panic(fmt.Sprintf("unknown type %s. Please declare it", v))
			}
			props[k] = Properties{
				Type:       Obj,
				Properties: p,
			}
		}
	}
	lookup[name] = props
	return s.WithProperties(props)
}

func BuildTypeGraph(t map[string]Type) *TypeGraph {
	ret := &TypeGraph{
		Nodes: map[string]TypeNode{
			"string": {
				Name: "string",
			},
			"int": {
				Name: "int",
			},
			"float": {
				Name: "float",
			},
			"bool": {
				Name: "bool",
			},
			"signal": {
				Name: "signal",
			},
			"time": {
				Name: "time",
			},
		},
	}
	for k, v := range t {
		ret.AddNode(NodeFromType(k, v))
	}
	return ret
}

func (t *TypeGraph) Properties(ty string) map[string]Properties {
	node, ok := t.Nodes[ty]
	if !ok {
		return nil
	}
	props := make(map[string]Properties)
	for _, v := range node.Fields {
		if TokenType(v.Type.Name).IsPrimitive() {
			props[v.Name] = Properties{
				Name: v.Name,
				Type: TokenType(v.Type.Name),
			}
			continue
		}
		if v.Type.IsArray {
			props[v.Name] = Properties{
				Name: v.Type.Name,
				Type: Arr,
				Properties: map[string]Properties{
					"items": {
						Type: TokenType(v.Type.Name),
					},
				},
			}
		} else {
			props[v.Name] = Properties{
				Name:       v.Type.Name,
				Type:       Obj,
				Properties: t.Properties(v.Type.Name),
			}
		}
	}
	return props
}

func (t *TypeGraph) Schema(ty string) *TokenSchema {
	return NewTokenSchema(ty).WithProperties(t.Properties(ty))
}

func (t *TypeGraph) Types() []*TokenSchema {
	ret := make([]*TokenSchema, 0)
	for k := range t.Nodes {
		schema := t.Schema(k)
		if schema.Type.IsPrimitive() {
			continue
		}
		ret = append(ret, schema)
	}
	return ret
}

func NetTypes(n *Net) map[string]Type {
	ret := make(map[string]Type)
	for _, schema := range n.TokenSchemas {
		if schema.Type.IsPrimitive() {
			continue
		}
		props := schema.Properties
		if props == nil {
			continue
		}
		typ := make(map[string]string)
		for k, v := range props {
			if v.Type.IsPrimitive() {
				typ[k] = string(v.Type)
				continue
			}
			if v.Type == Arr {
				typ[k] = fmt.Sprintf("[]%s", v.Properties["items"].Type)
				continue
			}
			typ[k] = v.Name
		}
		ret[schema.Name] = typ
	}
	return ret
}
