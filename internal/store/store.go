// Package store defines how links are persisted. Milestone 1 ships an
// in-memory implementation; a PostgreSQL one is the next step.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a short code does not exist.
var ErrNotFound = errors.New("link not found")

// Link is a shortened URL.
type Link struct {
	ID        uint64
	Code      string
	URL       string
	CreatedAt time.Time
}

// Store is the persistence boundary. Handlers depend on this interface only,
// so implementations (memory, Postgres, ...) can be swapped freely.
type Store interface {
	Create(ctx context.Context, url string) (Link, error)
	Get(ctx context.Context, code string) (Link, error)
}
