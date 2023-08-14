package main

import (
	"context"
	"embed"
	"github.com/jt05610/petri/control"
	"github.com/jt05610/petri/labeled"
)

//go:embed net.json
var netJSON embed.FS

//go:embed eventMap.json
var eventMapJSON embed.FS

type Valve struct {
}

func (v *Valve) OpenA(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *Valve) OpenB(ctx context.Context, event *labeled.Event) (*labeled.Event, error) {
	return event, nil
}

func (v *Valve) Handlers() control.Handlers {
	return control.Handlers{
		"open_a": v.OpenA,
		"open_b": v.OpenB,
	}
}
