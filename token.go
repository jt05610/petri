package petri

import (
	"errors"
	"fmt"
)

var _ Object = (*TokenSchema)(nil)
var _ Input = (*TokenInput)(nil)
var _ Update = (*TokenUpdate)(nil)
var _ Filter = (*TokenFilter)(nil)

type TokenType string

type PropertiesInput map[string]Properties

func (p PropertiesInput) Properties() *Properties {
	properties := make(map[string]Properties)
	for key, value := range p {
		properties[key] = value
	}
	return &Properties{
		Type:       Obj,
		Properties: properties,
	}
}

type TokenInput struct {
	Name       string
	Type       TokenType
	Properties PropertiesInput
}

func (t *TokenInput) Object() Object {
	obj := &TokenSchema{
		ID:         ID(),
		Name:       t.Name,
		Type:       t.Type,
		Properties: t.Properties,
	}
	return obj
}

func (t *TokenInput) Kind() Kind {
	return TokenObject
}

type TokenUpdate struct {
	Name string
	Type string
	Mask TokenMask
}

type TokenMask struct {
	Name bool
	Type bool
}

type Selector[T any] struct {
	Equals              T `json:"$eq,omitempty"`
	GreaterThan         T `json:"$gt,omitempty"`
	GreaterThanOrEquals T `json:"$gte,omitempty"`
	LessThan            T `json:"$lt,omitempty"`
	LessThanOrEquals    T `json:"$lte,omitempty"`
}

type TokenFilter struct {
	Name       *Selector[string]     `json:"name,omitempty"`
	Type       *Selector[TokenType]  `json:"type,omitempty"`
	Properties *Selector[Properties] `json:"properties,omitempty"`
}

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
		Type: String,
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
		Type: Boolean,
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
	Float   = TokenType("float")
	Int     = TokenType("integer")
	String  = TokenType("string")
	Boolean = TokenType("boolean")
	Obj     = TokenType("object")
)

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

func (t *TokenSchema) From(doc Document) error {
	//TODO implement me
	panic("implement me")
}

func (t *TokenSchema) String() string {
	return t.Name
}

// Token is an instance of a TokenSchema.
type Token struct {
	// Schema is the TokenSchema that this token is an instance of.
	Schema *TokenSchema
	// Value is the value of the token.
	Value interface{}
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
		return String, nil
	case "boolean":
		return Boolean, nil
	case "object":
		return Obj, nil
	default:
		return "", errors.New("invalid token type")
	}
}

func (t *TokenSchema) Update(update Update) error {
	upd, ok := update.(*TokenUpdate)
	if !ok {
		return errors.New("invalid update type")
	}
	if upd.Mask.Type {
		tokType, err := parseTokenType(upd.Type)
		if err != nil {
			return err
		}
		t.Type = tokType
	}
	if upd.Mask.Name {
		t.Name = upd.Name
	}
	return nil
}
func (t *Token) Copy() *Token {
	return &Token{
		Schema: t.Schema,
		Value:  t.Value,
	}
}

func (t *Token) String() string {
	return fmt.Sprintf("%s(%v)", t.Schema.Name, t.Value)
}

type InvalidTokenValueError struct {
	TokenSchema *TokenSchema
	Value       interface{}
}

func (e *InvalidTokenValueError) Error() string {
	return "invalid value for token " + e.TokenSchema.Name + ": " + e.Value.(string)
}

// NewToken creates a new token from the schema.
func (t *TokenSchema) NewToken(value interface{}) (*Token, error) {
	return &Token{
		Schema: t,
		Value:  value,
	}, nil
}

// WithValue sets the value of the token.
func (t *Token) WithValue(value interface{}) *Token {
	t.Value = value
	return t
}

type Handler interface {
	Handle(token *Token) (*Token, error)
}

type Generator interface {
	Handler
	Generate(value interface{}) (*Token, error)
}

type Transformer interface {
	Handler
	Transform(token *Token) (*Token, error)
}

type Consumer interface {
	Handler
	Consume(token *Token) error
}

type generator struct {
	f func(value interface{}) (*Token, error)
}

func (g *generator) Generate(value interface{}) (*Token, error) {
	return g.f(value)
}

func (g *generator) Handle(token *Token) (*Token, error) {
	return g.f(token.Value)
}

func NewGenerator(f func(value interface{}) (*Token, error)) Generator {
	return &generator{
		f: f,
	}
}

type transformer struct {
	f func(token *Token) (*Token, error)
}

func (t *transformer) Handle(token *Token) (*Token, error) {
	return t.f(token)
}

func (t *transformer) Transform(token *Token) (*Token, error) {
	return t.f(token)
}

func NewTransformer(f func(token *Token) (*Token, error)) Transformer {
	return &transformer{
		f: f,
	}
}

type consumer struct {
	f func(token *Token) error
}

func (c *consumer) Consume(token *Token) error {
	return c.f(token)
}

func (c *consumer) Handle(token *Token) (*Token, error) {
	return nil, c.f(token)
}

func NewConsumer(f func(token *Token) error) Consumer {
	return &consumer{
		f: f,
	}
}

func (t *TokenInput) IsInput()   {}
func (t *TokenUpdate) IsUpdate() {}
func (t *TokenFilter) IsFilter() {}
