package petri

import "sync"

// FIFO is a first-in-first-out queue. It can optionally be limited in size. If it is limited, then when the queue is
// full, the oldest items will be removed to make room for new items.
type FIFO[T any] struct {
	Values []T
	limit  int
	mu     sync.Mutex
}

func (fifo *FIFO[T]) Push(value ...T) {
	fifo.mu.Lock()
	defer fifo.mu.Unlock()
	if fifo.limit > 0 && len(fifo.Values)+len(value) > fifo.limit {
		fifo.Values = fifo.Values[len(value):]
	}
	fifo.Values = append(fifo.Values, value...)
}

func (fifo *FIFO[T]) Pop(n int) []T {
	fifo.mu.Lock()
	defer fifo.mu.Unlock()
	if n > len(fifo.Values) {
		n = len(fifo.Values)
	}
	values := fifo.Values[:n]
	fifo.Values = fifo.Values[n:]
	return values
}

func (fifo *FIFO[T]) Len() int {
	fifo.mu.Lock()
	defer fifo.mu.Unlock()
	return len(fifo.Values)
}

func NewFIFO[T any](limit int) *FIFO[T] {
	return &FIFO[T]{
		limit:  limit,
		Values: make([]T, 0, limit),
	}
}
