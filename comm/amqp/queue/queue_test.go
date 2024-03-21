package queue_test

import (
	"context"
	"errors"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp/queue"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"testing"
	"time"
)

func TestLocal_Enqueue(t *testing.T) {
	// Make sure that when we enqueue a token locally, we can see it happen with a remote queue. The remote should see
	// the token in the queue without removing it.
	exchange := "test_exchange"
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Error(err)
	}
	replyConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	ch, err := conn.Channel()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := ch.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	schema := petri.String()
	pl := petri.NewPlace("place", 1, schema)
	pl.ID = "test_place"
	local := queue.NewLocal(exchange, ch, pl)
	go func() {
		err := local.Serve(context.Background())
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(250 * time.Millisecond)
	tok, err := schema.NewToken([]byte("token"))
	if err != nil {
		t.Error(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err = local.Enqueue(ctx, tok)
	if err != nil {
		t.Error(err)
	}
	// make sure that newly declared remotes see the token
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch, err := replyConn.Channel()
			if err != nil {
				t.Error(err)
				return
			}
			remote := queue.NewRemote(exchange, ch, pl)
			tokens, err := remote.Peek(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			if len(tokens) != 1 {
				t.Errorf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].String() != "token" {
				t.Errorf("expected token, got %s", tokens[0].String())
			}
		}()
	}
	wg.Wait()
	// and make sure that those remotes don't remove the token
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	actual, err := local.Peek(ctx)
	if err != nil {
		t.Error(err)
	}
	if len(actual) != 1 {
		t.Errorf("expected 1 token, got %d", len(actual))
	}
	if actual[0].String() != "token" {
		t.Errorf("expected token, got %s", actual[0].String())
	}
	// have all the remotes try to dequeue the token and make sure only one can
	tokResults := make(map[int]bool)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ch, err := conn.Channel()
			if err != nil {
				t.Error(err)
				return
			}
			defer func() {
				err := ch.Close()
				if err != nil {
					t.Error(err)
				}
			}()
			remote := queue.NewRemote(exchange, ch, pl)
			tok, err := remote.Dequeue(ctx)
			if err != nil {
				if !errors.Is(err, queue.NoToken) {
					t.Error(err)
				}
			}
			if tok != nil {
				tokResults[i] = true
			}
		}(i)
	}
	wg.Wait()
	if len(tokResults) != 1 {
		t.Errorf("expected 1 remote to dequeue, got %d", len(tokResults))
	}
}
