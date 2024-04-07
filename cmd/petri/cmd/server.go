/*
Copyright © 2024 Jonathan Taylor <jonrtaylor12@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/caser"
	api "github.com/jt05610/petri/v1"
	"os"
	"slices"
	"strings"
	"sync"
	"text/template"

	"github.com/spf13/cobra"
)

type Language interface {
	GenerateServer(parDir string, n *petri.Net, remoteNets []string, g *DependencyGraph) error
}

var language string

var languageOptions = []string{"go", "typescript", "python", "ts", "py"}

type Go struct {
	packageName string
	typeDeps    []string
	tidy        bool
}

func doOrFail(err error) {
	if err != nil {
		panic(err)
	}
}

func ignoreReturn[T any](_ T, err error) {
	doOrFail(err)
}

func returnOrFail[T any](t T, err error) T {
	doOrFail(err)
	return t
}

func (g Go) initModule(parDir string, name string) {
	doOrFail(os.MkdirAll(fmt.Sprintf("%s/%s", parDir, name), os.ModePerm))
	if _, err := os.Stat(fmt.Sprintf("%s/go.mod", parDir)); err == nil {
		return
	}
	ignoreReturn(cmdExec(fmt.Sprintf("cd %s && go mod init %s", parDir, name)))
}

func (g Go) writeImports(imports []string, extra ...string) string {
	ret := "import (\n"
	ret += fmt.Sprintf("    \"github.com/jt05610/petri\"\n")
	ret += fmt.Sprintf("    api \"github.com/jt05610/petri/v1\"\n")
	for _, e := range extra {
		ret += fmt.Sprintf("    \"%s\"\n", e)

	}

	for _, imp := range imports {
		ret += fmt.Sprintf("    \"%s/%s/proto/v1/%s\"\n", g.packageName, imp, imp)
	}
	ret += ")\n"
	return ret

}

func stringIn(s string, ss []string) bool {
	for _, s1 := range ss {
		if s == s1 {
			return true
		}
	}
	return false
}

func typeImports(netName string, ts map[string]*petri.TokenSchema) []string {
	ret := make([]string, 0)
	seen := make(map[string]bool)
	for _, t := range ts {
		if stringIn(strings.ToLower(t.Name), []string{"time", "signal", "float", "bool", "string", "int"}) {
			continue
		}
		tName := t.Name
		if !strings.Contains(tName, ".") {
			if _, found := seen[tName]; found {
				continue
			}
			ret = append(ret, netName)
			seen[netName] = true
		}
		name := strings.Split(tName, ".")[0]
		if _, found := seen[name]; found {
			continue
		}
		seen[tName] = true
		ret = append(ret, name)
	}
	return ret
}

func shouldWrite(pkgName string, seen map[string]bool, ts *petri.TokenSchema) (string, bool) {
	if stringIn(strings.ToLower(ts.Name), []string{"time", "signal", "float", "bool", "string", "int"}) {
		return "", false
	}
	var pkg, name string
	if strings.Contains(ts.Name, ".") {
		pkg, name = strings.Split(ts.Name, ".")[0], strings.Split(ts.Name, ".")[1]
		if stringIn(strings.ToLower(name), []string{"time", "signal", "float", "bool", "string", "int"}) {
			return "", false
		}
		return ts.Name, true
	} else {
		if ts.Package != pkgName {
			return "", false
		}
		pkg = caser.New(ts.Package).CamelCase()
		name = ts.Name
	}
	tName := fmt.Sprintf("%s.%s", pkg, name)
	if _, found := seen[tName]; found {
		return "", false
	}
	seen[ts.Name] = true
	return ts.Name, true
}

func (g Go) writeTypeMap(types []*petri.TokenSchema) string {
	ret := "func TypeMap(n *petri.Net) api.TokenTypeMap{\n"
	ret += "    return api.TokenTypeMap{\n"
	seen := make(map[string]bool)
	for _, t := range types {
		if name, ok := shouldWrite(g.packageName, seen, t); ok {
			ret += fmt.Sprintf("    n.TokenSchema(\"%s\"): new(%s).ProtoReflect().Type(),\n", t, name)
		}
	}

	ret += "    }\n}\n"
	return ret
}

func (g Go) writeTypeFile(parDir string, n *petri.Net, types ...*petri.TokenSchema) {
	f := returnOrFail(os.Create(fmt.Sprintf("%s/%s/types.go", parDir, n.Name)))
	defer func() {
		doOrFail(f.Close())
	}()
	ignoreReturn(f.WriteString(fmt.Sprintf("package main\n\n")))

	ignoreReturn(f.WriteString(g.writeImports(typeImports(n.Name, n.TokenSchemas))))
	ignoreReturn(f.WriteString(g.writeTypeMap(types)))
}

func (g Go) writeClientsFile(parDir string, n *petri.Net, remoteNets []string) {
	f := returnOrFail(os.Create(fmt.Sprintf("%s/%s/clients.go", parDir, n.Name)))
	defer func() {
		doOrFail(f.Close())
	}()
	ignoreReturn(f.WriteString(fmt.Sprintf("package main\n\n")))
	netNames := []string{n.Name}
	for _, net := range remoteNets {
		netNames = append(netNames, strings.Split(net, ".")[0])
	}

	ignoreReturn(f.WriteString(g.writeImports(netNames, "google.golang.org/grpc")))
	for _, sn := range n.Nets {
		ignoreReturn(f.WriteString(g.writeSubnetHandlers(sn)))
	}
	ignoreReturn(f.WriteString(g.writeClient(n)))
}

func (g Go) writeService(n *petri.Net) string {
	ff := make([]struct {
		Name    string
		Place   string
		Package string
	}, 0)
	for t, tr := range n.Transitions {
		outputs := n.Inputs(tr)
		outPl := outputs[0].Place
		if strings.Contains(t, ".") {
			continue
		}
		ff = append(ff, struct {
			Name    string
			Place   string
			Package string
		}{
			Name:    caser.New(t).PascalCase(),
			Place:   outPl.ID,
			Package: g.packageName,
		})
	}
	data := struct {
		Package   string
		Service   string
		Functions []struct {
			Name    string
			Place   string
			Package string
		}
	}{
		Functions: ff,
		Package:   g.packageName,
		Service:   caser.New(n.Name).PascalCase(),
	}
	tpl := `
var _ {{.Package}}.{{.Service}}ServiceServer = (*service)(nil)

type service struct {
	*petri.Net
	{{.Package}}.UnimplementedDbtlServiceServer
	petri.Marking
	mu sync.RWMutex
}


{{ range .Functions}}func (s *service) {{.Name}}(ctx context.Context, request *{{.Package}}.{{.Name}}Request) (*{{.Package}}.{{.Name}}Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var resp {{.Package}}.{{.Name}}Response
	m, t, err := api.Handle[*{{.Package}}.{{.Name}}Request, *dbtl.{{.Name}}Response](s.Net, ctx, s.Marking, "{{.Place}}", request, resp.ProtoReflect().Type())
	if err != nil {
		return nil, err
	}
	s.Marking = m
	return t, nil
}

{{end}}
`
	return writeTemplate(g, data, tpl)
}

func (g Go) writeServiceFile(parDir string, n *petri.Net) {
	f := returnOrFail(os.Create(fmt.Sprintf("%s/%s/service.go", parDir, n.Name)))
	defer func() {
		doOrFail(f.Close())
	}()
	ignoreReturn(f.WriteString(fmt.Sprintf("package main\n\n")))
	ignoreReturn(f.WriteString(g.writeImports([]string{n.Name}, "context", "sync")))
	ignoreReturn(f.WriteString(g.writeService(n)))
}

func (g Go) GenerateServer(parDir string, n *petri.Net, remoteNets []string, dg *DependencyGraph) error {
	g.packageName = caser.New(n.Name).CamelCase()
	g.initModule(parDir, caser.New(n.Name).CamelCase())
	g.writeTypeFile(parDir, n, dg.Types()...)
	g.writeClientsFile(parDir, n, remoteNets)
	g.writeServiceFile(parDir, n)
	if g.tidy {
		tidyModule(parDir)
	}
	return nil
}

func executeTemplate(tpl *template.Template, data interface{}) string {
	var buf strings.Builder
	err := tpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type TransitionData struct {
	Route      string
	Transition string
}

func isCapital(s string) bool {
	return s[0] >= 'A' && s[0] <= 'Z'
}

func MakeTransitionData(n *petri.Net) []*TransitionData {
	ret := make([]*TransitionData, 0)
	for tn, t := range n.Transitions {
		if len(n.Inputs(t)) == 0 {
			continue
		}
		if isCapital(tn) {
			continue
		}
		ret = append(ret, &TransitionData{
			Route:      fmt.Sprintf("%s.%s", n.Name, tn),
			Transition: caser.New(t.Name).PascalCase(),
		})
	}
	slices.SortFunc(ret, func(i, j *TransitionData) int {
		return strings.Compare(i.Route, j.Route)
	})
	return ret
}

func writeTemplate[T any](g Go, data T, tpl string) string {
	t := template.Must(template.New("server").Parse(tpl))
	return executeTemplate(t, data)
}

func (g Go) writeSubnetHandlers(subNet *petri.Net) string {
	data := struct {
		ImportName  string
		PascalName  string
		Transitions []*TransitionData
	}{
		ImportName:  caser.New(subNet.Name).CamelCase(),
		PascalName:  caser.New(subNet.Name).PascalCase(),
		Transitions: MakeTransitionData(subNet),
	}
	tpl := `
func {{.ImportName}}TransitionMap(c {{.ImportName}}.{{.PascalName}}ServiceClient, optSrv *api.OptionService) api.HandlerMap {
	return api.HandlerMap{
{{range .Transitions}}		"{{.Route}}": api.NewHandler(api.TransitionClient(c.{{.Transition}}, optSrv)),	
{{end}}
	}
}
`
	return writeTemplate(g, data, tpl)
}

type GoClientData struct {
	Package string
	Service string
}

type GoClientTemplate struct {
	Clients []GoClientData
}

func LoadClientTemplateData(n *petri.Net) GoClientTemplate {
	ret := GoClientTemplate{}
	for _, sn := range n.Nets {
		ret.Clients = append(ret.Clients, GoClientData{
			Package: caser.New(sn.Name).CamelCase(),
			Service: caser.New(sn.Name).PascalCase(),
		})
	}
	return ret
}

func (g Go) writeClient(n *petri.Net) string {
	data := LoadClientTemplateData(n)
	tpl := `
type Client struct {
{{range .Clients}}	{{.Service}} {{.Package}}.{{.Service}}ServiceClient
{{end}}
	*api.OptionService
}

type Route struct {
	Name    string
	Options []grpc.DialOption
}

type ClientRoutes struct {
	Designer Route
	Builder  Route
	Tester   Route
	Learner  Route
}

func ConnectClient[T any](route Route, f func(connInterface grpc.ClientConnInterface) T) T {
	conn, err := grpc.Dial(route.Name, route.Options...)
	if err != nil {
		panic(err)
	}
	return f(conn)
}

func NewClient(routes *ClientRoutes) *Client {
	c := &Client{
		OptionService: api.NewOptionService(),
	}
{{range .Clients}}		c.{{.Service}} = ConnectClient(routes.{{.Service}}, {{.Package}}.New{{.Service}}ServiceClient)
{{end}}
	return c
}

func (c *Client) HandlerMap() api.HandlerMap {
	return api.MergeMaps(
{{range .Clients}}		{{.Package}}TransitionMap(c.{{.Service}}, c.OptionService),
{{end}}
	)
}`
	return writeTemplate(g, data, tpl)
}

type TypeScript struct {
}

var _ Language = (*Go)(nil)
var _ Language = (*TypeScript)(nil)
var _ Language = (*Python)(nil)

func (t TypeScript) GenerateServer(parDir string, n *petri.Net, remoteNets []string, dg *DependencyGraph) error {
	//TODO implement me
	panic("implement	me")
}

type Python struct {
}

func (p Python) GenerateServer(parDir string, n *petri.Net, remoteNets []string, dg *DependencyGraph) error {
	//TODO implement me
	panic("implement me")

}

func NewLanguage(lang string, tidy bool) Language {
	switch lang {
	case "go":
		return Go{
			tidy: tidy,
		}
	case "typescript", "ts":
		return TypeScript{}
	case "python", "py":
		return Python{}
	default:
		panic(fmt.Sprintf("unknown language %s", lang))
	}
}

func GenServer(parDir, in, lang string, remoteNets []string, tidy bool) {
	l := NewLanguage(lang, tidy)
	n := api.LoadNet(in)
	tree := NewDependencyGraph(n)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		err := l.GenerateServer(parDir, n, remoteNets, tree)
		if err != nil {
			panic(err)
		}
	}()
	for _, n := range tree.GetNets() {
		wg.Add(1)
		go func(n *petri.Net) {
			defer wg.Done()
			genProto(n, parDir)
		}(n)
	}
	wg.Wait()
	for _, n := range tree.GetNets() {
		wg.Add(1)
		go func(n *petri.Net) {
			defer wg.Done()
			goProtoClient(n, parDir)
		}(n)
	}
	wg.Wait()
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "generate a server",
	Long:  `generate a server for the petri net`,
	Run: func(cmd *cobra.Command, args []string) {
		GenServer(outputDir, inputFile, language, remotes, tidy)
	},
}

var remotes []string
var tidy bool

func init() {
	genCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&language, "language", "l", "go", "the language to generate the server in")
	serverCmd.Flags().StringSliceVarP(&remotes, "remotes", "r", []string{}, "remote subnets")
	serverCmd.Flags().BoolVarP(&tidy, "tidy", "t", true, "tidy up the generated code")
}
