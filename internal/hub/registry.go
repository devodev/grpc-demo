package hub

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// IDGenerator is an interface used to generate unique IDs.
type IDGenerator interface {
	GetID() uint64
}

// SimpleIDGenerator generates IDs by incrementing a local counter indefinitely.
type SimpleIDGenerator struct {
	counter uint64
}

// GetID implements the IDGenerator interface.
func (g *SimpleIDGenerator) GetID() uint64 {
	return atomic.AddUint64(&g.counter, 1)
}

// ClientRegistry is an interface used to store and retrieve clients.
type ClientRegistry interface {
	Get(uint64) (*Client, error)
	GetWithName(string) (*Client, error)
	Register(*Client) uint64
	Unregister(uint64)
	Count() int
}

// RegistryMem is an in-memory ClientRegistry.
type RegistryMem struct {
	idGenerator IDGenerator

	mu              *sync.Mutex
	clients         map[uint64]*Client
	clientNameIndex map[string]uint64
}

// NewRegistryMem returns an initialized RegistryMem.
func NewRegistryMem() *RegistryMem {
	return &RegistryMem{
		idGenerator:     new(SimpleIDGenerator),
		mu:              &sync.Mutex{},
		clients:         make(map[uint64]*Client),
		clientNameIndex: make(map[string]uint64),
	}
}

// Register implements the ClientRegistry interface.
// It uses the underlying IDGenerator for generating a new ID
// and returns it.
func (r *RegistryMem) Register(c *Client) uint64 {
	clientID := r.idGenerator.GetID()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[clientID] = c
	r.clientNameIndex[c.Name] = clientID
	return clientID
}

// Unregister implements the ClientRegistry interface.
// It removes the client assigned to the provided id.
func (r *RegistryMem) Unregister(id uint64) {
	if c, ok := r.clients[id]; ok {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.clientNameIndex, c.Name)
		delete(r.clients, id)
	}
}

// Get implements the ClientRegistry interface.
// It returns the client assigned to the provided id
// or a not found error otherwise.
func (r *RegistryMem) Get(id uint64) (*Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clients[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

// GetWithName .
func (r *RegistryMem) GetWithName(name string) (*Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clientNameIndex[name]
	if !ok {
		return nil, fmt.Errorf("client with name: %v not found", name)
	}
	return r.clients[c], nil
}

// Count implements the ClientRegistry interface.
// It returns the count of currently registered clients.
func (r *RegistryMem) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.clients)
}
