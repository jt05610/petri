package yaml_test

import (
	"context"
	"github.com/jt05610/petri/device/yaml"
	"os"
	"testing"
)

func TestService_Load(t *testing.T) {
	ctx, can := context.WithCancel(context.Background())
	defer can()
	dir := "../../petrifile/v1/yaml/examples"
	path := "../examples/light.yaml"

	t.Run("Local", func(t *testing.T) {
		srv := yaml.NewService(dir)
		in, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		dev, err := srv.Load(ctx, in)
		if err != nil {
			t.Fatal(err)
		}
		if dev == nil {
			t.Fatal("nil device")
		}
		if dev.Net == nil {
			t.Fatal("nil net")
		}
		for tn, tr := range dev.Net.Transitions {
			t.Logf("%s: %v", tn, tr)
		}
		err = dev.Connect("amqp://guest:guest@localhost:5672/")
		if err != nil {
			t.Fatal(err)
		}
		dev.Close()
	})

	t.Run("LocalWithRemote", func(t *testing.T) {
		path := "../examples/light_switch_logger.yaml"
		srv := yaml.NewService(dir)
		in, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		dev, err := srv.Load(ctx, in)
		if err != nil {
			t.Fatal(err)
		}
		if dev == nil {
			t.Fatal("nil device")
		}
		if dev.Net == nil {
			t.Fatal("nil net")
		}
		for tn, tr := range dev.Net.Transitions {
			t.Logf("%s: %v", tn, tr)
		}
		err = dev.Connect("amqp://guest:guest@localhost:5672/")
		if err != nil {
			t.Fatal(err)
		}
		dev.Close()
	})
}
