package protobuf

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/caser"
	"io"
	"slices"
	"sort"
	"strings"
)

type Service struct {
}

// Field represents the data structure for declaring a protobuf field
type Field struct {
	Name     caser.Caser
	Type     string
	Number   int
	Optional bool
	Repeated bool
}

func (f Field) String() string {
	if f.Optional {
		return fmt.Sprintf("optional %s %s = %d", f.Type, f.Name.SnakeCase(), f.Number)
	}
	if f.Repeated {
		return fmt.Sprintf("repeated %s %s = %d", f.Type, f.Name.SnakeCase(), f.Number)
	}
	return fmt.Sprintf("%s %s = %d", f.Type, f.Name.SnakeCase(), f.Number)
}

// Message represents the data structure for declaring a protobuf message
type Message struct {
	Name   caser.Caser
	Fields []Field
}

func (m Message) String() string {
	s := fmt.Sprintf("message %s {\n", m.Name.PascalCase())
	for _, f := range m.Fields {
		s += fmt.Sprintf("  %s;\n", f)
	}
	return s + "}\n"
}

type RPC struct {
	Name   string
	Input  string
	Output string
}

func (r RPC) String() string {
	return fmt.Sprintf("  rpc %s (%s) returns (%s);", caser.New(r.Name).PascalCase(), r.Input, r.Output)
}

type Data struct {
	Parent   string
	Name     caser.Caser
	Messages []Message
	RPCs     []RPC
	Imports  []string
}

func (d Data) AddMessage(name string, fields []Field) Data {
	m := Message{
		Name:   caser.New(name),
		Fields: fields,
	}
	d.Messages = append(d.Messages, m)
	slices.SortFunc(d.Messages, func(a, b Message) int {
		return strings.Compare(a.Name.SnakeCase(), b.Name.SnakeCase())
	})

	return d
}

func (d Data) AddRPC(name, input, output string) Data {
	r := RPC{
		Name:   name,
		Input:  input,
		Output: output,
	}
	d.RPCs = append(d.RPCs, r)
	slices.SortFunc(d.RPCs, func(a, b RPC) int {
		return strings.Compare(a.Name, b.Name)
	})
	return d
}

func (d Data) AddImport(imp string) Data {
	d.Imports = append(d.Imports, imp)
	return d
}

func NewData(name string, parent ...string) Data {
	if len(parent) > 0 {
		return Data{
			Parent:   parent[0],
			Name:     caser.New(name),
			Messages: make([]Message, 0),
			RPCs:     make([]RPC, 0),
			Imports:  make([]string, 0),
		}
	}
	return Data{
		Parent:   name,
		Name:     caser.New(name),
		Messages: make([]Message, 0),
		RPCs:     make([]RPC, 0),
		Imports:  make([]string, 0),
	}
}

func (d Data) String() string {
	s := "syntax = \"proto3\";\n\n"
	s += fmt.Sprintf("package %s;\n\n", d.Name.SnakeCase())
	s += fmt.Sprintf("option go_package = \"%s/proto/v1/%s\";\n\n", caser.New(d.Parent).CamelCase(), d.Name.CamelCase())
	sort.Strings(d.Imports)
	for _, imp := range d.Imports {
		s += fmt.Sprintf("import \"%s/proto/v1/%s\";\n", strings.Split(imp, ".")[0], imp)
	}
	s += "\n"
	seenMessages := make(map[string]bool)
	for i, m := range d.Messages {
		if _, ok := seenMessages[m.Name.SnakeCase()]; ok {
			continue
		}
		seenMessages[m.Name.SnakeCase()] = true
		s += m.String()
		if i < len(d.Messages)-1 {
			s += "\n"
		}
	}
	if len(d.RPCs) == 0 {
		return s

	}
	s += fmt.Sprintf("\nservice %sService {\n", d.Name.PascalCase())
	seenRPCs := make(map[string]bool)
	for _, r := range d.RPCs {
		if _, ok := seenRPCs[caser.ToPascalCase(r.Name)]; ok {
			continue
		}
		seenRPCs[caser.ToPascalCase(r.Name)] = true
		s += r.String() + "\n"
	}
	s += "}\n"
	return s
}

func IsInputEvent(n petri.Net, t *petri.Transition) bool {
	return len(n.Inputs(t)) == 0
}

type Crawler struct {
	Net     *petri.Net
	Visited map[string]bool
}

var TypeMap = map[petri.TokenType]string{
	petri.Str:       "string",
	petri.Int:       "int32",
	petri.Float:     "float",
	petri.Sig:       "google.protobuf.Empty",
	petri.Bool:      "bool",
	petri.TimeStamp: "google.protobuf.Timestamp",
}

func SortStrings(ss []string) []string {
	sort.Strings(ss)
	return ss
}

func Sorted[T any](m map[string]T) []struct {
	Name  string
	Value T
} {
	ret := make([]struct {
		Name  string
		Value T
	}, 0, len(m))
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	names = SortStrings(names)
	for _, n := range names {
		ret = append(ret, struct {
			Name  string
			Value T
		}{Name: n, Value: m[n]})
	}
	return ret
}

func MakeInputTransitions(net *petri.Net) *petri.Net {
	for _, pl := range net.InputPlaces() {
		name := strings.Split(caser.New(pl.Name).PascalCase(), "Input")[0]
		tr := petri.NewTransition(name)
		net = net.WithTransitions(tr)
		net = net.WithArcs(petri.NewArc(pl, tr, pl.AcceptedTokens[0].Name, pl.AcceptedTokens[0]))
	}
	return net
}

func MakeOutputTransitions(net *petri.Net) *petri.Net {
	for _, pl := range net.OutputPlaces() {
		name := strings.Split(caser.New(pl.Name).PascalCase(), "Input")[0]
		tr := petri.NewTransition(name)
		net = net.WithTransitions(tr)
		net = net.WithArcs(petri.NewArc(tr, pl, pl.AcceptedTokens[0].Name, pl.AcceptedTokens[0]))
	}
	return net
}

func CrawlTokenSchema(net *petri.Net) []Message {
	types := petri.NetTypes(net)
	ret := make([]Message, 0, len(types))
	for _, ty := range Sorted(types) {
		n := ty.Name
		t := ty.Value
		if petri.TokenType(n).IsPrimitive() || strings.Contains(n, ".") {
			continue
		}
		msg := Message{
			Name:   caser.New(n),
			Fields: make([]Field, 0),
		}
		for _, typeDecl := range Sorted(t) {
			fName := typeDecl.Name
			fType := typeDecl.Value
			if petri.TokenType(fType).IsPrimitive() {
				msg.Fields = append(msg.Fields, Field{
					Name:     caser.New(fName),
					Type:     TypeMap[petri.TokenType(fType)],
					Number:   len(msg.Fields) + 1,
					Optional: true,
				})
				continue
			}
			msg.Fields = append(msg.Fields, Field{
				Name:     caser.New(fName),
				Type:     petri.ArrayOf(fType),
				Number:   len(msg.Fields) + 1,
				Optional: !petri.IsArray(fType),
				Repeated: petri.IsArray(fType),
			})
		}
		ret = append(ret, msg)
	}
	return ret
}

func NetData(net *petri.Net, parentDir string) Data {
	net = MakeInputTransitions(net)
	net = MakeOutputTransitions(net)
	d := NewData(caser.New(net.Name).PascalCase(), parentDir)
	imports := make([]string, 0)
	seenImports := make(map[string]bool)
	for _, t := range net.Transitions {
		if !net.Owns(t) {
			continue
		}
		fields := make([]Field, 0)
		outputFields := make([]Field, 0)
		i := 0
		for _, pl := range net.Inputs(t) {
			if pl.Place.AcceptedTokens[0].Name == "signal" {
				continue
			}
			if strings.Contains(pl.Place.AcceptedTokens[0].Name, ".") {
				seenImports[pl.Place.AcceptedTokens[0].Name] = true
				splName := strings.Split(pl.Place.AcceptedTokens[0].Name, ".")
				if _, ok := seenImports[strings.Join(splName[0:len(splName)-1], ".")]; !ok {
					imports = append(imports, strings.Join(splName[0:len(splName)-1], "."))
					seenImports[strings.Join(splName[0:len(splName)-1], ".")] = true
				}
			}
			fName := pl.Place.Name

			if strings.Contains(fName, "Input") {
				fName = "input"
			}

			fields = append(fields, Field{
				Name:     caser.New(fName),
				Type:     pl.Place.AcceptedTokens[0].Name,
				Number:   i + 1,
				Optional: true,
			})
			i++
		}
		i = 0
		for _, pl := range net.Outputs(t) {
			if pl.Place.AcceptedTokens[0].Name == "signal" {
				continue
			}
			outputFields = append(outputFields, Field{
				Name:     caser.New(pl.Place.Name),
				Type:     pl.Place.AcceptedTokens[0].Name,
				Number:   i + 1,
				Optional: true,
			})
			i++

		}
		d = d.AddMessage(fmt.Sprintf("%sResponse", caser.New(t.Name).PascalCase()), outputFields)
		d = d.AddMessage(fmt.Sprintf("%sRequest", caser.New(t.Name).PascalCase()), fields)
		d = d.AddRPC(t.Name, fmt.Sprintf("%sRequest", caser.New(t.Name).PascalCase()), fmt.Sprintf("%sResponse", caser.New(t.Name).PascalCase()))
	}
	msgs := CrawlTokenSchema(net)
	for _, m := range msgs {
		d = d.AddMessage(m.Name.PascalCase(), m.Fields)
	}
	writtenImports := make(map[string]bool)
	for _, imp := range imports {
		if _, ok := writtenImports[imp]; !ok {
			d = d.AddImport(fmt.Sprintf("%s.proto", imp))
			writtenImports[imp] = true
		}
	}

	return d
}

func (s *Service) Flush(_ context.Context, w io.Writer, n *petri.Net, parentDir string) error {
	_, err := w.Write([]byte(NetData(n, parentDir).String()))
	return err
}
