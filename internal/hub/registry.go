package hub

import (
	"fmt"
	"sync"
)

// ClientRegistry is an interface used to store and retrieve clients.
type ClientRegistry interface {
	Get(string) (*Client, error)
	Register(*Client, string) error
	Unregister(string) error
	Count() int
}

// RegistryMem is an in-memory ClientRegistry.
type RegistryMem struct {
	mu      *sync.Mutex
	clients map[string]*Client
}

// NewRegistryMem returns an initialized RegistryMem.
func NewRegistryMem() *RegistryMem {
	return &RegistryMem{
		mu:      &sync.Mutex{},
		clients: make(map[string]*Client),
	}
}

// Register implements the ClientRegistry interface.
// It registers a client using the name provided.
func (r *RegistryMem) Register(c *Client, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clients[name]; ok {
		return fmt.Errorf("registration failed because client Name %v already exists", name)
	}
	r.clients[name] = c
	return nil
}

// Unregister implements the ClientRegistry interface.
// It removes the client assigned to the provided name.
func (r *RegistryMem) Unregister(name string) error {
	if _, ok := r.clients[name]; !ok {
		return fmt.Errorf("unregistration failed because client Name was not found")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, name)
	return nil
}

// Get implements the ClientRegistry interface.
// It returns the client assigned to the provided name.
func (r *RegistryMem) Get(name string) (*Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clients[name]
	if !ok {
		return nil, fmt.Errorf("get failed because client Name was not found")
	}
	return c, nil
}

// Count implements the ClientRegistry interface.
// It returns the count of currently registered clients.
func (r *RegistryMem) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.clients)
}
