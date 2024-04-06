package cmd

import (
	"testing"
)

func TestGo_GenerateServer(t *testing.T) {
	outdir := "test"
	language = "go"
	nName := "../../../examples/dbtl/petri/dbtl.yaml"
	GenServer(outdir, nName, language)
}
