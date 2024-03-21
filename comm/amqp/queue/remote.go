package queue

import (
	"context"
	"fmt"
	"github.com/jt05610/petri"
	amqp "github.com/rabbitmq/amqp091-go"
)

var _ petri.TokenQueue = (*Remote)(nil)

// Remote is a token queue belonging to another petri.Net. All token operations must be performed remotely.
type Remote struct {
	*Queue
}

// NewRemote creates a new Remote queue.
func NewRemote(exchange string, ch *amqp.Channel, pl *petri.Place) *Remote {
	q := NewQueue(ch, pl.AcceptedTokens[0], exchange, pl.ID)
	return &Remote{
		Queue: q,
	}
}

// Enqueue adds tokens to the queue.
func (r *Remote) Enqueue(ctx context.Context, tt ...*petri.Token) error {
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

func (r *Remote) Dequeue(ctx context.Context) (*petri.Token, error) {
	ch, err := r.rpc(ctx, "get")
	if err != nil {
		return nil, err
	}
	tok := <-ch
	if tok == nil {
		return nil, NoToken
	}
	err = r.pop(ctx, tok)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func (r *Remote) Copy() petri.TokenQueue {
	return &Remote{
		Queue: r.Queue,
	}
}

func (r *Remote) Channel(ctx context.Context) <-chan *petri.Token {
	panic("implement me")
}

func (r *Remote) Available(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Remote) Peek(ctx context.Context) ([]*petri.Token, error) {
	tokens := make([]*petri.Token, 0)
	ch, err := r.rpc(ctx, "peek")
	if err != nil {
		return nil, err
	}
	for t := range ch {
		tokens = append(tokens, t)
	}
	return tokens, nil
}
