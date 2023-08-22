package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
)

type TwoPositionThreeWayValve struct {
	*os.File
	mu sync.Mutex
}

func (d *TwoPositionThreeWayValve) do(c cmd) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, b := range c.Bytes() {
		_, err := d.Write(b)
		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(d)
		if err != nil {
			panic(err)
		}
		if strings.Trim(buf.String(), "\r\n") != "ok" {
			return fmt.Errorf("error: %v", buf.String())
		}
	}
	return nil
}

type cmd string

// Bytes splits the command by line and returns slices of bytes for each line.
func (c cmd) Bytes() [][]byte {
	split := strings.Split(string(c), "\n")
	b := make([][]byte, len(split))
	for i, s := range split {
		b[i] = []byte(s)
	}
	return b
}

const (
	OpenA cmd = "M5"
	OpenB cmd = "M3"
)

type OpenARequest struct {
}

type OpenAResponse struct {
}

type OpenBRequest struct {
}

type OpenBResponse struct {
}
