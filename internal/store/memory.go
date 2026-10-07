package store

import (
	"context"
	"sync"
	"time"

	"shortener/internal/base62"
)

// startID keeps the first codes from being a single character long.
const startID = 100000

// Memory is a concurrency-safe in-memory Store. Data is lost on restart.
type Memory struct {
	mu    sync.RWMutex
	next  uint64
	links map[string]Link
}

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{next: startID, links: make(map[string]Link)}
}

// Create stores url and returns the new Link with a generated code.
func (m *Memory) Create(_ context.Context, url string) (Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.next
	m.next++

	link := Link{
		ID:        id,
		Code:      base62.Encode(id),
		URL:       url,
		CreatedAt: time.Now().UTC(),
	}
	m.links[link.Code] = link
	return link, nil
}

// Get looks a link up by its short code.
func (m *Memory) Get(_ context.Context, code string) (Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	link, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return link, nil
}
