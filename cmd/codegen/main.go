package main

import "github.com/jt05610/petri/cmd/codegen/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
