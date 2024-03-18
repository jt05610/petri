package main

import (
	"context"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/device/yaml"
	"os"
)

var dev *device.Device

func doOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func retOrPanic[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	doOrPanic(dev.Run(context.Background()))
}

func init() {
	srv := yaml.NewService("desc")
	f := retOrPanic(os.Open("desc/light.yaml"))
	dev = retOrPanic(srv.Load(nil, f))
	doOrPanic(dev.Connect("amqp://guest:guest@localhost:5672"))
}
