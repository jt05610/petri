package petri

type Document map[string]interface{}

type Object interface {
	Kind() Kind
	Identifier() string
	String() string
	Update(update Update) error
	Document() Document
	From(doc Document) error
}

type Kind int

const (
	PlaceObject Kind = iota
	TransitionObject
	ArcObject
	NetObject
	TokenObject
)
