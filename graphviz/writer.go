package graphviz

import (
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/jt05610/petri"
	"io"
)

var _ petri.Flusher[*petri.Net] = (*Writer)(nil)

type Writer struct {
	*Config
	g       *cgraph.Graph
	mapping map[petri.Node]*cgraph.Node
}

func (w *Writer) writePlace(i int, p *petri.Place) error {
	name := fmt.Sprintf("p%d", i)
	node, err := w.g.CreateNode(name)
	if err != nil {
		return err
	}
	node.SetShape(cgraph.CircleShape)
	node.SetLabel(p.Name)
	node.Set("fontname", string(w.Font))
	w.mapping[p] = node
	return nil
}

func (w *Writer) writeTransition(i int, t *petri.Transition) error {
	name := fmt.Sprintf("t%d", i)
	node, err := w.g.CreateNode(name)
	if err != nil {
		return err
	}
	w.mapping[t] = node
	node.SetShape(cgraph.BoxShape)
	node.SetLabel(t.Name)
	node.Set("fontname", string(w.Font))
	return nil
}

func (w *Writer) writeArc(i int, a *petri.Arc) error {
	src := w.mapping[a.Src]
	dst := w.mapping[a.Dest]
	name := fmt.Sprintf("a%d", i)
	_, err := w.g.CreateEdge(name, src, dst)
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) Flush(out io.Writer, t *petri.Net) error {
	graph := graphviz.New()
	defer func() {
		_ = graph.Close()
	}()
	g, err := graph.Graph()
	if err != nil {
		return err
	}
	g.SetRankDir(cgraph.RankDir(w.RankDir))
	w.g = g
	for i, p := range t.Places {
		if err := w.writePlace(i, p); err != nil {
			return err
		}
	}
	for i, t := range t.Transitions {
		if err := w.writeTransition(i, t); err != nil {
			return err
		}
	}
	for i, a := range t.Arcs {
		if err := w.writeArc(i, a); err != nil {
			return err
		}
	}
	if err := graph.Render(w.g, graphviz.XDOT, out); err != nil {
		return err
	}
	return nil
}

type Font string

func (f Font) Or(other Font) Font {
	return f + "," + other
}

const (
	Helvetica  Font = "Helvetica"
	Arial      Font = "Arial"
	Roboto     Font = "Roboto"
	Montserrat Font = "Montserrat"
	SansSerif  Font = "sans-serif"
	Serif      Font = "Serif"
	Times      Font = "Times"
)

type RankDir string

const (
	LeftToRight RankDir = "LR"
	RightToLeft RankDir = "RL"
	TopToBottom RankDir = "TB"
	BottomToTop RankDir = "BT"
)

type Config struct {
	Name string
	Font
	RankDir
}

func New(config *Config) *Writer {
	if config.Name == "" {
		config.Name = "petri"
	}
	return &Writer{
		Config:  config,
		mapping: make(map[petri.Node]*cgraph.Node),
	}
}
