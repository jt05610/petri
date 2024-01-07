package main

import (
	"embed"
	"net/http"
)

//go:embed public/*
var public embed.FS

func main() {
	addr := ":50000"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(public)))
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}

}
