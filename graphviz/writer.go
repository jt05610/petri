package graphviz

import (
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/jt05610/petri"
	"io"
	"strings"
)

var _ petri.Flusher[*petri.Net] = (*Writer)(nil)

type Writer struct {
	*Config
	g       *cgraph.Graph
	mapping map[petri.Node]*cgraph.Node
	legend  *cgraph.Graph
}

func mapToGraphvizRecord(m map[string]interface{}) string {
	var fields []string
	i := 0
	for k, v := range m {
		// For simplicity, we're converting values directly to strings.
		// You might need more sophisticated handling for nested maps or other complex types.
		fields = append(fields, fmt.Sprintf("<f%d> %s | %v", i, k, v))
		i++
	}
	// Join all fields with a separator to form the label of the record node
	recordLabel := strings.Join(fields, " | ")

	return fmt.Sprintf("digraph G {\n    node [shape=record];\n    \"node1\" [label=\"{%s}\"];\n}", recordLabel)
}

func crawlProperties(m map[string]petri.Properties) map[string]interface{} {
	r := make(map[string]interface{})
	for k, v := range m {
		if v.Properties != nil {
			val := make(map[string]interface{})
			for k, v := range crawlProperties(v.Properties) {
				val[k] = v
			}
			r[k] = val
			continue
		}
		r[k] = v.Type
	}
	return r
}

func Join[T any](tt []T, sep string) string {
	var b strings.Builder
	for i, t := range tt {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(fmt.Sprintf("%v", t))
	}
	return b.String()
}

func WriteMapAsRecord(b *strings.Builder, m map[string]interface{}) error {
	kk := make([]string, 0, len(m)+1)
	vv := make([]interface{}, 0, len(m)+1)
	kk = append(kk, "Field")
	vv = append(vv, "Type")
	for k, v := range m {
		kk = append(kk, k)
		vv = append(vv, v)
	}
	_, err := b.WriteString(fmt.Sprintf("{%s}|", strings.Join(kk, "|")))
	if err != nil {
		return err
	}
	_, err = b.WriteString(fmt.Sprintf("{%v}", Join(vv, "|")))

	return err
}

func (w *Writer) WriteTokenSchema(name string, ts *petri.TokenSchema) error {
	node, err := w.legend.CreateNode(name)
	if err != nil {
		return err
	}
	bld := new(strings.Builder)
	if ts.Properties != nil {
		bld.WriteString("{{" + ts.Name + "}|{")
		err = WriteMapAsRecord(bld, crawlProperties(ts.Properties))
		if err != nil {
			return err
		}
		bld.WriteString("}}")
	} else {
		bld.WriteString(fmt.Sprintf("{%s | %v}", ts.Name, ts.Type))

	}

	node.SetShape("record")
	node.SetLabel(bld.String())
	node.Set("fontname", string(w.Font))
	node.Set("labeljust", "l")
	return nil
}

func (w *Writer) writePlace(i string, p *petri.Place) error {
	name := fmt.Sprintf("p_%s", i)
	node, err := w.g.CreateNode(name)
	if err != nil {
		return err
	}
	node.SetShape(cgraph.EllipseShape)
	node.SetLabel(p.Name)
	node.Set("fontname", string(w.Font))
	w.mapping[p] = node
	return nil
}

func (w *Writer) writeTransition(i string, t *petri.Transition) error {
	name := fmt.Sprintf("t_%s", i)
	node, err := w.g.CreateNode(name)
	if err != nil {
		return err
	}
	w.mapping[t] = node
	node.SetShape(cgraph.BoxShape)
	node.SetLabel(t.Name)
	node.Set("fontname", string(w.Font))
	if t.Expression != "" {
		node.Set("labeljust", "l")
		node.SetLabel(fmt.Sprintf("%s\n%s", t.Name, t.Expression))
	}
	return nil
}

func (w *Writer) writeArc(i int, a *petri.Arc) error {
	src := w.mapping[a.Src]
	dst := w.mapping[a.Dest]
	name := fmt.Sprintf("a%d", i)
	edge, err := w.g.CreateEdge(name, src, dst)
	if err != nil {
		return err
	}
	edge.SetLabel(fmt.Sprintf("%s", a.Expression))
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
	g.SetNewRank(true)
	w.g = g.SubGraph(t.Name, 1)
	w.g.SetLabel(t.Name)
	w.g.SetRankDir(cgraph.RankDir(w.RankDir))
	w.g.Set("rank", "same")

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
	w.legend = g.SubGraph("cluster_legend", 1)
	if err != nil {
		return err
	}
	w.legend.SetLabel("Legend")
	w.legend.Set("rank", "same")
	w.legend.SetLabelJust("l")
	w.legend.SetLabelLocation(cgraph.TopLocation)
	w.legend.SetStyle(cgraph.FilledGraphStyle)
	w.legend.SetRankDir(cgraph.LRRank)
	for i, a := range t.TokenSchemas {
		if err := w.WriteTokenSchema(i, a); err != nil {
			return err
		}
	}
	if err := graph.Render(g, graphviz.SVG, out); err != nil {
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
