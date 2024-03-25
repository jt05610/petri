package queue

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
)

var _ petri.TokenQueue = (*Remote)(nil)

// Remote is a token queue belonging to another petri.Net. All token operations must be performed remotely.
type Remote struct {
	*Queue
}

func (r *Remote) Close() {
	//TODO implement me
	panic("implement me")
}

// NewRemote creates a new Remote queue.
func NewRemote(exchange string, ch *amqp.Channel, pl *petri.Place) *Remote {
	q := NewQueue(ch, pl.AcceptedTokens[0], exchange, pl.ID)
	return &Remote{
		Queue: q,
	}
}

// Enqueue adds tokens to the queue.
func (r *Remote) Enqueue(ctx context.Context, tt ...petri.Token) error {
	for _, t := range tt {
		ch, err := r.rpc(ctx, "post", t)
		if err != nil {
			return err
		}
		<-ch
		err = r.put(ctx, t)
		if err != nil {
			return err
		}
	}
	return nil
}

var NoToken = fmt.Errorf("no token")

func (r *Remote) Dequeue(ctx context.Context) (petri.Token, error) {
	ch, err := r.rpc(ctx, "get")
	var zero petri.Token
	if err != nil {
		return zero, err
	}
	tok := <-ch
	if tok.Value == nil {
		return zero, NoToken
	}
	err = r.pop(ctx, tok)
	if err != nil {
		return zero, err
	}
	return tok, nil
}

func (r *Remote) Copy() petri.TokenQueue {
	return &Remote{
		Queue: r.Queue,
	}
}

func (r *Remote) Monitor(ctx context.Context) <-chan []petri.Token {
	// make listeners on the remotes in and out topics. When we get something from the in channel, we add it to a queue and send it. We remove when we get something from the out channel
	ch := make(chan []petri.Token)
	sub, err := r.subscribe(ctx)
	if err != nil {
		panic(err)
	}
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case tok := <-sub:
				if tok.op == "in" {
					ch <- []petri.Token{tok.Token}
				} else {
					ch <- []petri.Token{}
				}
			}
		}
	}()
	return ch
}

func (r *Remote) Available(ctx context.Context) (int, error) {
	ch, err := r.rpc(ctx, "available")
	if err != nil {
		return 0, err
	}
	v := <-ch
	if v.Value == nil {
		return 0, nil
	}
	valStr := string(v.Bytes())
	iVal, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, err
	}
	return iVal, nil
}

func (r *Remote) Peek(ctx context.Context) ([]petri.Token, error) {
	tokens := make([]petri.Token, 0)
	ch, err := r.rpc(ctx, "peek")
	if err != nil {
		return nil, err
	}
	for t := range ch {
		tokens = append(tokens, t)
	}
	return tokens, nil
}
