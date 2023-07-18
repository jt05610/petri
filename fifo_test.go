package petri_test

import (
	"petri"
	"sync"
	"testing"
)

func TestFIFO_Concurrency(t *testing.T) {
	var wg sync.WaitGroup
	f := petri.NewFIFO[int](0)
	concurrent := 1000
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				f.Push(i)
				f.Pop(1)
			}
		}()
	}
	wg.Wait()
}
