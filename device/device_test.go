package device_test

import (
	"context"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp"
	"github.com/jt05610/petri/device"
	"github.com/jt05610/petri/examples"
	"github.com/rabbitmq/amqp091-go"
	"testing"
	"time"
)

func TestDevice_Run(t *testing.T) {
	// This test makes sure that a device can create a connection with a currently running device.
	// 1. Launch the external device.
	external := device.Device{Net: examples.Switch(), Initial: map[string]string{"off": "signal"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := external.Connect(ctx, "amqp://guest:guest@localhost:5672")
		if err != nil {
			panic(err)

		}
		defer external.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	// 2. Load the device that depends on the external device.
	d := device.Device{
		Net:     examples.LightSwitch(),
		Initial: map[string]string{"light.off": "signal", "switch.off": "signal"},
		Remotes: map[string]string{"switch": "amqp://guest:guest@localhost:5672"},
	}

	// 3. Run the device.
	go func() {
		err := d.Connect(ctx, "amqp://guest:guest@localhost:5672")
		if err != nil {
			t.Error(err)
			return
		}
		defer d.Close()
		err = d.Run(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	// 5. Make a client for the dependent device and make a call that sends a signal to the external device.
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		t.Fatal(err)
	}
	client, err := amqp.NewClient(conn, "lightSwitch", 500*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	schema := petri.Signal()
	tok, err := schema.NewToken([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1000 * time.Millisecond)
	original := d.State()
	t.Log("got initial state", original)
	err = client.Put(ctx, &amqp.PutRequest{Place: "switch.turnOn", Token: tok.Bytes()})
	if err != nil {
		t.Fatal(err)
	}
	updated := d.State()

	// 6. Check that the dependent device has received the signal.
	if amqp.SameMarking(original, updated) {
		t.Fatal("markings are equal")
	}
	t.Log("external marking")
	// 7. Check that the external device has received the signal.
	externalState := external.State()
	if !amqp.SameMarking(updated, externalState) {
		t.Fatal("markings are not equal")
	}
}
