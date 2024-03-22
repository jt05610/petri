package petri

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// TokenQueue is a queue of tokens currently in a place.
type TokenQueue interface {
	// Enqueue adds tokens to the queue.
	Enqueue(ctx context.Context, tt ...Token) error
	// Dequeue removes a token from the queue.
	Dequeue(ctx context.Context) (Token, error)
	// Copy returns a copy of the queue.
	Copy() TokenQueue
	// Monitor returns a channel that will send the queue's current marking whenever it changes.
	Monitor(ctx context.Context) <-chan []Token
	// Available returns the number of tokens in the queue.
	Available(ctx context.Context) (int, error)
	// Peek returns the tokens in the queue without removing them.
	Peek(ctx context.Context) ([]Token, error)
}

type LocalQueue struct {
	tokens  []Token
	max     int
	updated chan struct{}
	mu      sync.Mutex
}

func (l *LocalQueue) Peek(_ context.Context) ([]Token, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tokens, nil
}

func (l *LocalQueue) Enqueue(_ context.Context, tt ...Token) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.tokens)+len(tt) > l.max {
		return errors.New("queue is full")
	}
	l.tokens = append(l.tokens, tt...)
	select {
	case l.updated <- struct{}{}:
	default:
	}
	return nil
}

func (l *LocalQueue) Dequeue(_ context.Context) (Token, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var zero Token
	if len(l.tokens) == 0 {
		return zero, errors.New("queue is empty")
	}
	token := l.tokens[0]
	l.tokens = l.tokens[1:]
	select {
	case l.updated <- struct{}{}:
	default:
	}
	return token, nil
}

func (l *LocalQueue) Copy() TokenQueue {
	return &LocalQueue{
		tokens:  l.tokens,
		max:     l.max,
		updated: make(chan struct{}),
	}
}

func (l *LocalQueue) Monitor(ctx context.Context) <-chan []Token {
	ch := make(chan []Token)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
			return
		case <-l.updated:
			t, _ := l.Peek(ctx)
			ch <- t
		}
	}()
	return ch
}

func (l *LocalQueue) Available(_ context.Context) (int, error) {
	return len(l.tokens), nil
}

var _ TokenQueue = (*LocalQueue)(nil)

func NewLocalQueue(max int) *LocalQueue {
	return &LocalQueue{
		max:     max,
		tokens:  make([]Token, 0, max),
		updated: make(chan struct{}),
	}
}

func (l *LocalQueue) String() string {
	return fmt.Sprintf("%v", l.tokens)
}
