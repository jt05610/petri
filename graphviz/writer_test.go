package graphviz_test

import (
	"fmt"
	"github.com/jt05610/petri/examples"
	"github.com/jt05610/petri/graphviz"
	"os"
	"os/exec"
	"testing"
)

func TestWriter_Flush(t *testing.T) {
	net := examples.Net()
	cfg := &graphviz.Config{
		Font:    graphviz.Helvetica,
		RankDir: graphviz.LeftToRight,
	}
	df, err := os.Create("net.dot")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = df.Close()
	}()
	w := graphviz.New(cfg)
	err = w.Flush(df, net)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range net.Arcs {
		fmt.Println(a)
	}
	cmdStr := "dot net.dot -Tsvg -Nfontname=Helvetica > net.svg"
	cmd := exec.Command("sh", "-c", cmdStr)
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
