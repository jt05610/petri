package codegen

import (
	"embed"
	"github.com/jt05610/petri/labeled"
)

//go:embed template
var templateDir embed.FS

type Language string

type Gen struct {
	Language
	InstanceID string
	Port       string
	Events     []*labeled.Event
}
