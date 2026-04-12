package memory

import (
	"context"
	"errors"
)

var (
	ErrInvalidInput = errors.New("memory: invalid input")
	ErrNotFound     = errors.New("memory: not found")
)

// Repository is the storage boundary the rest of the application depends on.
// Concrete backends such as SQLite satisfy this contract.
//
// Implementations should honor context cancellation, validate caller input,
// and avoid retaining references to caller-owned mutable data.
type Repository interface {
	// Create validates and persists input, then returns the canonical saved
	// memory including system-assigned fields such as ID and timestamps. It
	// returns ErrInvalidInput when the input cannot be accepted.
	Create(ctx context.Context, input CreateInput) (Memory, error)

	// Get returns the canonical saved memory for id. It returns ErrInvalidInput
	// when id is malformed and ErrNotFound when no matching memory exists.
	Get(ctx context.Context, id string) (Memory, error)
}
