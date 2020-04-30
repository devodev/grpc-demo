package hub

import (
	"sync"
	"sync/atomic"
)

type registry struct {
	counter uint64

	mu      *sync.Mutex
	clients map[uint64]*client
}

func newRegistry() *registry {
	return &registry{
		mu:      &sync.Mutex{},
		clients: make(map[uint64]*client),
	}
}

func (r *registry) getID() uint64 {
	return atomic.AddUint64(&r.counter, 1)
}

func (r *registry) registerClient(c *client) uint64 {
	clientID := r.getID()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[clientID] = c
	return clientID
}

func (r *registry) unregisterClient(id uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, id)
}

func (r *registry) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.clients)
}
