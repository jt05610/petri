package petri

import (
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"testing"
)

func TestProtobufService_Load(t *testing.T) {
	n := LoadNet("../examples/dbtl/petri/dbtl.yaml")
	if n == nil {
		t.Fatal("failed to load net")
	}

	fset := token.NewFileSet()
	src := `
package main
	
import (
	"github.com/jt05610/petri"
	model "github.com/jt05610/petri/v1/testModel"
)
	var TypeMap = petri.TokenTypeMap{
		n.TokenSchema("model.Sample"): new(model.Sample).ProtoReflect().Type(),
	}

`
	f, err := parser.ParseFile(fset, "", src, parser.Trace)
	if err != nil {
		t.Fatal(err)
	}
	err = printer.Fprint(os.Stdout, fset, f)
	if err != nil {
		t.Fatal(err)
	}
}
