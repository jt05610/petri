package amqp

import (
	"context"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"testing"
	"time"
)

const URL = "amqp://guest:guest@localhost:5672/"

const Timeout = 1 * time.Second

func TestServer_Connect(t *testing.T) {
	conn, err := amqp.Dial(URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()
	n := petri.NewNet("test")
	server := NewServer(conn, n, nil).WithTimeout(Timeout)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := server.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	client, err := NewClient(conn, n.Name, Timeout)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Connect(ctx, &ConnectRequest{
		Url:            URL,
		Exchange:       "",
		PlaceName:      "",
		TransitionName: "",
		Expression:     "",
		Direction:      0,
	})
	if err != nil {
		t.Fatal(err)
	}
}
