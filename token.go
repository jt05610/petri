package petri

import (
	"errors"
	"fmt"
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
	return "integer"
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
	Int       = TokenType("integer")
	Str       = TokenType("string")
	Bool      = TokenType("boolean")
	Obj       = TokenType("object")
	Sig       = TokenType("signal")
	TimeStamp = TokenType("time")
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
	Value
}

func (t *TokenSchema) CanAccept(fields []string) bool {
	if t.Type != Obj {
		return false
	}
	fieldIdx := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		fieldIdx[field] = struct{}{}
	}

	met := make(map[string]bool, len(t.Properties))
	for key := range t.Properties {
		_, met[key] = fieldIdx[key]
	}
	for f, _ := range fieldIdx {
		if !met[f] {
			fmt.Printf("field %s not found\n", f)
			return false
		}
	}
	return true
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

func (t *TokenSchema) PropertiesJSON() map[string]interface{} {
	if t.Properties != nil {
		ret := make(map[string]interface{})
		for key, value := range t.Properties {
			ret[key] = value
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

type Value interface {
	Bytes() []byte
	FromBytes([]byte) error
}

// Token is an instance of a TokenSchema.
type Token struct {
	// ID is the unique identifier of the token.
	ID string `json:"_id"`
	// Schema is the schema of the token.
	Schema *TokenSchema `json:"schema"`
	// Value is the value of the token.
	Value `json:"value"`
}

func (t *Token) String() string {
	return fmt.Sprintf("%s(%v)", t.Schema.Name, t.Value)
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
func (t *TokenSchema) NewToken(value []byte) (*Token, error) {
	tok := &Token{
		ID:     ID(),
		Schema: t,
		Value:  t.Value,
	}
	var err error
	if tok.Value != nil {
		err = tok.Value.FromBytes(value)
	}
	if err != nil {
		return nil, err
	}
	return tok, nil
}

type Handler interface {
	Handle(token ...*Token) ([]*Token, error)
}

type stringValue struct {
	value string
}

func (s stringValue) Bytes() []byte {
	return []byte(s.value)
}

func (s stringValue) FromBytes(bytes []byte) error {
	s.value = string(bytes)
	return nil
}

var (
	_ Value = (*stringValue)(nil)
	_ Value = (*signalValue)(nil)
)

type signalValue struct {
}

func (s signalValue) Bytes() []byte {
	return []byte{1}
}

func (s signalValue) FromBytes(b []byte) error {
	if b == nil {
		return errors.New("nil signal")
	}
	return nil
}

func Signal() *TokenSchema {
	return &TokenSchema{
		ID:    ID(),
		Name:  "signal",
		Type:  Sig,
		Value: signalValue{},
	}
}

func String() *TokenSchema {
	return &TokenSchema{
		ID:    ID(),
		Name:  "string",
		Type:  Str,
		Value: stringValue{},
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
