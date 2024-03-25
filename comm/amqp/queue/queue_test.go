package queue_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri"
	"github.com/jt05610/petri/comm/amqp/queue"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"testing"
	"time"
)

const exchange = "test_exchange"
const nRemotes = 1

func TestLocal_Enqueue(t *testing.T) {
	// Make sure that when we enqueue a token locally, we can see it happen with a remote queue. The remote should see
	// the token in the queue without removing it.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	go func() {
		err := local.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	tok, err := schema.NewToken([]byte("token"))
	if err != nil {
		t.Error(err)
	}
	defer cancel()
	err = local.Enqueue(ctx, tok)
	if err != nil {
		t.Error(err)
	}
	// make sure that newly declared remotes see the token
	var wg sync.WaitGroup
	for i := 0; i < nRemotes; i++ {
		wg.Add(1)
		go func() {
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
	for i := 0; i < nRemotes; i++ {
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
			available, err := remote.Available(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			if available != 1 {
				return
			}
			tok, err := remote.Dequeue(ctx)
			if err != nil {
				if !errors.Is(err, queue.NoToken) {
					t.Error(err)
				}
			}
			if tok.Value != nil {
				tokResults[i] = true
			}
		}(i)
	}
	wg.Wait()
	if len(tokResults) != 1 {
		t.Errorf("expected 1 remote to dequeue, got %d", len(tokResults))
	}
}

func TestRemote_Monitor(t *testing.T) {
	// want to make sure that we are getting marking updates from remote places so that we can handle them accordingly
	// Make sure that when we enqueue a token locally, we can see it happen with a remote queue. The remote should see
	// the token in the queue without removing it.
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
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
	ctx, can := context.WithCancel(context.Background())
	defer can()
	history := make([][]string, nRemotes)
	updateCh := make(chan []petri.Token)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case upd := <-updateCh:
				fmt.Printf("got local update: %v\n", upd)
			}
		}
	}()
	go func() {
		err := local.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	tokenValues := []string{
		"one",
		"two",
		"three",
		"four",
		"five",
	}
	var wg sync.WaitGroup
	for i := 0; i < nRemotes; i++ {
		remote := queue.NewRemote(exchange, ch, pl)
		monitor := remote.Monitor(ctx)
		wg.Add(1)
		go func(i int) {
			history[i] = make([]string, 0, len(tokenValues)*2)
			defer func() {
				if len(history[i]) != len(tokenValues)*2 {
					t.Errorf("expected %d updates, got %d", len(tokenValues)*2, len(history[i]))
				}

				for j, expected := range tokenValues {
					h := history[i][j*2]
					if h != expected {
						t.Errorf("expected %s, got %s", expected, h)
					}
				}
				wg.Done()
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
			for {
				select {
				case <-ctx.Done():
					return
				case upd := <-monitor:
					fmt.Printf("got update: %v\n", upd)
					if len(upd) != 1 {
						history[i] = append(history[i], "")
						continue
					}
					history[i] = append(history[i], upd[0].String())
				}
			}
		}(i)
	}
	time.Sleep(250 * time.Millisecond)
	for _, v := range tokenValues {
		tok, err := schema.NewToken([]byte(v))
		if err != nil {
			t.Error(err)
		}
		err = local.Enqueue(ctx, tok)
		if err != nil {
			t.Error(err)
		}
		_, err = local.Dequeue(ctx)
		if err != nil {
			panic(err)
		}
	}
	time.Sleep(100 * time.Millisecond)
	can()
	wg.Wait()

}
