package graphviz_test

import (
	"fmt"
	"github.com/jt05610/petri/examples"
	"github.com/jt05610/petri/graphviz"
	"os"
	"testing"
)

func TestWriter_Flush(t *testing.T) {
	net := examples.NewPump()
	cfg := &graphviz.Config{
		Font:    graphviz.Helvetica,
		RankDir: graphviz.LeftToRight,
	}
	df, err := os.Create("net.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = df.Close()
	}()
	w := graphviz.New(cfg)
	err = w.Flush(df, net.Net)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range net.Arcs {
		fmt.Println(a)
	}
}
