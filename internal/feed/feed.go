package feed

import (
	"fmt"
	"sync"
)

// Feed .
type Feed struct {
	ch chan string

	mu      *sync.Mutex
	readers map[chan string]struct{}
}

// New .
func New() *Feed {
	return &Feed{
		ch: make(chan string),

		mu:      &sync.Mutex{},
		readers: make(map[chan string]struct{}),
	}
}

// Send .
func (f *Feed) Send(message string) error {
	select {
	case f.ch <- message:
	default:
		return fmt.Errorf("channel closed")
	}
	return nil
}

// GetCh .
func (f *Feed) GetCh(quit chan struct{}) <-chan string {
	f.mu.Lock()
	defer f.mu.Unlock()
	ch := make(chan string)
	f.readers[ch] = struct{}{}

	go func() {
		defer delete(f.readers, ch)
		select {
		case <-quit:
		}
	}()

	return ch
}

// StartRouter .
func (f *Feed) StartRouter(quit chan struct{}) {
	var wg sync.WaitGroup
Loop:
	for {
		select {
		case <-quit:
			break Loop
		case message := <-f.ch:
			f.mu.Lock()
			wg.Add(len(f.readers))
			for readerCh := range f.readers {
				go func(ch chan string) {
					defer wg.Done()
					ch <- message
				}(readerCh)
			}
			wg.Wait()
			f.mu.Unlock()
		}
	}
	for ch := range f.readers {
		close(ch)
	}
}
