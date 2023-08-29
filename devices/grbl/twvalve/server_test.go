package main

import (
	"context"
	"github.com/jt05610/petri/amqp/server"
	"github.com/jt05610/petri/labeled"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"testing"
)

func loopback(ctx context.Context) (tx chan []byte, rx <-chan []byte) {
	rxCh := make(chan []byte)
	txCh := make(chan []byte)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-txCh:
				rxCh <- b
			}
		}
	}()
	return rxCh, txCh
}

func TestHandlers(t *testing.T) {
	ctx := context.Background()
	txCh, rxCh := loopback(ctx)
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	d := NewTwoPositionThreeWayValve(txCh, rxCh)
	h := d.Handlers()
	if h == nil {
		t.Error("expected non-nil handlers")
	}
	dev := d.load()
	env := LoadEnv(logger)
	conn, err := amqp.Dial(env.URI)
	failOnError(err, "Failed to connect to RabbitMQ")
	logger.Info("Connected to RabbitMQ", zap.String("uri", env.URI))
	defer func() {
		err := conn.Close()
		failOnError(err, "Failed to close connection")
	}()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	logger.Info("Opened channel")
	srv := server.New(dev.Nets[0], ch, env.Exchange, env.DeviceID, env.DeviceID, dev.EventMap(), d.Handlers())
	curMarking := srv.MarkingMap()
	go func() {
		ch := srv.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-ch:
				t.Logf("got %+v", b)
			}
		}
	}()
	err = srv.Handle(ctx, &labeled.Event{
		Name: "open_b",
	})

	if err != nil {
		t.Fatal(err)
	}
	newMarking := srv.MarkingMap()
	oneDiff := false
	for k, v := range newMarking {
		if curMarking[k] != v {
			oneDiff = true
		}
	}

	if !oneDiff {
		t.Error("expected at least one difference in marking")
	}
}
