package graphviz_test

import (
	"bytes"
	"github.com/jt05610/petri/examples"
	"github.com/jt05610/petri/graphviz"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	fn := "testdata/net.dot"
	df, err := os.Open(fn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = df.Close()
	}()
	ld := graphviz.Loader()
	_, err = ld.Load(df)
}

func TestE2E(t *testing.T) {
	buf := new(bytes.Buffer)
	cfg := &graphviz.Config{
		Name:    "",
		Font:    graphviz.Helvetica,
		RankDir: graphviz.LeftToRight,
	}

	wr := graphviz.New(cfg)
	net := examples.NewPump().Net
	err := wr.Flush(buf, net)
	if err != nil {
		t.Fatal(err)
	}
	ld := graphviz.Loader()
	read, err := ld.Load(buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Places) != len(net.Places) {
		t.Fatal("places mismatch")
	}
	if len(read.Transitions) != len(net.Transitions) {
		t.Fatal("transitions mismatch")
	}
	if len(read.Arcs) != len(net.Arcs) {
		t.Fatal("arcs mismatch")
	}
	for _, arc := range read.Arcs {
		if net.Arc(arc.Src, arc.Dest) == nil {
			t.Fatalf("arc mismatch: %s", arc)
		}
	}
}
