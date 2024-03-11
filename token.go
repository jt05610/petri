package petri

import (
	"errors"
	"fmt"
)

var _ Object = (*TokenSchema)(nil)
var _ Input = (*TokenSchemaInput)(nil)
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

type TokenSchemaInput struct {
	Name       string
	Type       TokenType
	Properties PropertiesInput
}

func (t *TokenSchemaInput) Object() Object {
	obj := &TokenSchema{
		ID:         ID(),
		Name:       t.Name,
		Type:       t.Type,
		Properties: t.Properties,
	}
	return obj
}

func (t *TokenSchemaInput) Kind() Kind {
	return TokenObject
}

type TokenUpdate struct {
	Name       string
	Type       string
	Properties PropertiesInput
	Mask       TokenMask
}

type TokenMask struct {
	Name bool
	Type bool
}

type TokenFilter struct {
	ID         *StringSelector       `json:"_id,omitempty"`
	Name       *StringSelector       `json:"name,omitempty"`
	Type       *StringSelector       `json:"type,omitempty"`
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
		_, ok := value.(int)
		return ok
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

// Token is an instance of a TokenSchema.
type Token struct {
	// ID is the unique identifier of the token.
	ID string `json:"_id"`
	// Schema is the schema of the token.
	Schema *TokenSchema `json:"schema"`
	// Value is the value of the token.
	Value interface{} `json:"value"`
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

type InvalidTokenValueError struct {
	TokenSchema *TokenSchema
	Value       interface{}
}

func (e *InvalidTokenValueError) Error() string {
	return fmt.Sprintf("invalid value for token %s: %v", e.TokenSchema.Name, e.Value)
}

// NewToken creates a new token from the schema.
func (t *TokenSchema) NewToken(value interface{}) (*Token, error) {
	return &Token{
		ID:     ID(),
		Schema: t,
		Value:  value,
	}, nil
}

type Handler interface {
	Handle(token ...*Token) ([]*Token, error)
}

type Generator interface {
	Handler
	Generate(value ...interface{}) ([]*Token, error)
}

type Transformer interface {
	Handler
	Transform(token ...*Token) ([]*Token, error)
}

type Consumer interface {
	Handler
	Consume(token ...*Token) error
}

type generator struct {
	f func(value ...interface{}) ([]*Token, error)
}

func (g *generator) Generate(values ...interface{}) ([]*Token, error) {
	return g.f(values...)
}

func (g *generator) Handle(tokens ...*Token) ([]*Token, error) {
	return g.Generate()
}

func NewGenerator(f func(value ...interface{}) ([]*Token, error)) Generator {
	return &generator{
		f: f,
	}
}

type transformer struct {
	f func(token ...*Token) ([]*Token, error)
}

func (t *transformer) Handle(token ...*Token) ([]*Token, error) {
	return t.f(token...)
}

func (t *transformer) Transform(tokens ...*Token) ([]*Token, error) {
	return t.f(tokens...)
}

func NewTransformer(f func(tokens ...*Token) ([]*Token, error)) Transformer {
	return &transformer{
		f: f,
	}
}

type consumer struct {
	f func(token ...*Token) error
}

func (c *consumer) Consume(token ...*Token) error {
	return c.f(token...)
}

func (c *consumer) Handle(tokens ...*Token) ([]*Token, error) {
	return nil, c.f(tokens...)
}

func NewConsumer(f func(token ...*Token) error) Consumer {
	return &consumer{
		f: f,
	}
}

type SignalToken struct {
	*TokenSchema
}

func (s *SignalToken) NewToken(value interface{}) (*Token, error) {
	return &Token{
		ID:     ID(),
		Schema: s.TokenSchema,
		Value:  1,
	}, nil
}

func Signal() *TokenSchema {
	return &TokenSchema{
		ID:   ID(),
		Name: "signal",
		Type: Sig,
	}
}

func (t *TokenSchemaInput) IsInput() {}
func (t *TokenUpdate) IsUpdate()     {}
func (t *TokenFilter) IsFilter()     {}

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
