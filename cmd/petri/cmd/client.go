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
	"github.com/jt05610/petri/protobuf"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Generator struct {
	OutDir string
}

func genProto(net *petri.Net, outDir string) {
	protoDir := outDir + "/" + net.Name + "/proto/v1"
	err := os.MkdirAll(protoDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	path := protoDir + "/" + caser.New(net.Name).SnakeCase() + ".proto"
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()
	s := protobuf.Service{}
	pathSplit := strings.Split(outDir, "/")
	pkg := pathSplit[len(pathSplit)-1] + "/" + net.Name
	err = s.Flush(nil, f, net, pkg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated proto file at: %s\n", path)
}

type Node struct {
	Net       *petri.Net
	DependsOn map[string]*Node
}

type DependencyGraph struct {
	Nodes map[string]*Node
}

func NewDependencyGraph(net *petri.Net) *DependencyGraph {
	g := &DependencyGraph{
		Nodes: map[string]*Node{},
	}
	g.buildTree(&Node{
		Net:       net,
		DependsOn: make(map[string]*Node),
	})
	return g
}

func (d *DependencyGraph) GetNets() []*petri.Net {
	nets := make([]*petri.Net, 0, len(d.Nodes))
	for _, node := range d.Nodes {
		nets = append(nets, node.Net)
	}
	return nets
}

func (d *DependencyGraph) buildTree(node *Node) {
	if _, seen := d.Nodes[node.Net.Name]; seen {
		return
	}
	d.Nodes[node.Net.Name] = node
	for _, subnet := range node.Net.Nets {
		var depNode *Node
		if _, ok := d.Nodes[subnet.Name]; !ok {
			depNode = &Node{
				Net:       subnet,
				DependsOn: make(map[string]*Node),
			}
		} else {
			depNode = d.Nodes[subnet.Name]
		}
		d.buildTree(depNode)
		node.DependsOn[subnet.Name] = depNode
	}

}

func initModule(dir string) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	name := strings.Split(dir, "/")[len(strings.Split(dir, "/"))-1]
	_, err = cmdExec(fmt.Sprintf("cd %s && go mod init %s", dir, name))
	if err != nil {
		panic(err)
	}
}

func tidyModule(dir string) {
	_, err := cmdExec(fmt.Sprintf("cd %s && go mod tidy", dir))
	if err != nil {
		panic(err)
	}
}

func genClient(net *petri.Net, outDir string) {
	tree := NewDependencyGraph(net)
	for _, n := range tree.GetNets() {
		genProto(n, outDir)
	}
	for _, n := range tree.GetNets() {
		goProtoClient(n, outDir)
	}
}

func cmdExec(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command %s: %s", command, string(out))
	}
	return string(out), nil
}

func goProtoClient(net *petri.Net, outDir string) {
	protoPath := fmt.Sprintf("%s/proto/v1/%s.proto", caser.New(net.Name).CamelCase(), caser.New(net.Name).SnakeCase())
	protoDir := outDir + "/" + caser.New(net.Name).CamelCase()
	goPath := strings.Split(protoPath, ".")[0]
	goOut := outDir + "/" + goPath
	err := os.MkdirAll(goOut, os.ModePerm)
	if err != nil {
		panic(err)
	}
	protoPaths := fmt.Sprintf("--proto_path=%s/proto/v1 --proto_path=%s", protoDir, outDir)
	for i := range net.Nets {
		protoPaths += fmt.Sprintf(" --proto_path=%s/proto/v1", outDir+"/"+caser.New(net.Nets[i].Name).CamelCase())
	}
	parentDir := filepath.Dir(outDir)
	if outDir == "." {
		parentDir = ".."
	}
	cmd := fmt.Sprintf("protoc --go_out=%s --go-grpc_out=%s %s %s.proto", parentDir, parentDir, protoPaths, caser.New(net.Name).SnakeCase())
	fmt.Println(cmd)
	_, err = cmdExec(cmd)
	if err != nil {
		panic(err)
	}
}

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		net := loadNet(inputFile)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			err := os.MkdirAll(outputDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		if _, err := os.Stat(outputDir + "/go.mod"); os.IsNotExist(err) {
			initModule(outputDir)
		}
		genClient(net, outputDir)
		tidyModule(outputDir)
	},
}

func init() {
	genCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
