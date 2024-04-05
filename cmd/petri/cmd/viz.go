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
	"github.com/jt05610/petri/graphviz"
	"github.com/spf13/cobra"
	"os"
)

var format string

// vizCmd represents the viz command
var vizCmd = &cobra.Command{
	Use:   "viz",
	Short: "Create a graphviz figure from a petri net",
	Long:  `Create a graphviz figure from a petri net. The input file must be a petri file.`,
	Run: func(cmd *cobra.Command, args []string) {
		net := loadNet(inputFile)
		outName := net.Name + "." + format
		cfg := &graphviz.Config{
			Name:    net.Name,
			Font:    graphviz.Helvetica,
			RankDir: graphviz.LeftToRight,
			Format:  graphviz.Format(format),
		}
		outPath := outputDir + "/" + outName
		fmt.Printf("writing figure for %s to %s...", inputFile, outPath)
		err := os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		df, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = df.Close()
		}()
		w := graphviz.New(cfg)
		err = w.Flush(df, net)
		if err != nil {
			panic(err)
		}
		fmt.Println("done")
	},
}

func init() {
	rootCmd.AddCommand(vizCmd)
	vizCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "input file")
	vizCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "output directory")
	vizCmd.PersistentFlags().StringVarP(&format, "format", "f", "svg", "output format")
}
