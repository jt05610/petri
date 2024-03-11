package petri

import (
	"errors"
	"fmt"
)

type TokenQueue interface {
	Enqueue(...*Token) error
	Dequeue() (*Token, error)
	Copy() TokenQueue
	Channel() <-chan *Token
	Available() int
}

type LocalQueue struct {
	tokens []*Token
	max    int
}

func (l *LocalQueue) Enqueue(token ...*Token) error {
	if len(l.tokens)+len(token) > l.max {
		return errors.New("queue is full")
	}
	l.tokens = append(l.tokens, token...)
	return nil
}

func (l *LocalQueue) Dequeue() (*Token, error) {
	if len(l.tokens) == 0 {
		return nil, errors.New("queue is empty")
	}
	token := l.tokens[0]
	l.tokens = l.tokens[1:]
	return token, nil
}

func (l *LocalQueue) Copy() TokenQueue {
	return &LocalQueue{
		tokens: l.tokens,
		max:    l.max,
	}
}

func (l *LocalQueue) Channel() <-chan *Token {
	ch := make(chan *Token)
	go func() {
		defer close(ch)
		for _, token := range l.tokens {
			ch <- token
		}
	}()
	return ch
}

func (l *LocalQueue) Available() int {
	return len(l.tokens)
}

var _ TokenQueue = (*LocalQueue)(nil)

func NewLocalQueue(max int) *LocalQueue {
	return &LocalQueue{
		max:    max,
		tokens: make([]*Token, 0, max),
	}
}

func (l *LocalQueue) String() string {
	return fmt.Sprintf("%v", l.tokens)
}
