package memory

import (
	"context"
	"errors"
)

var (
	ErrInvalidInput = errors.New("memory: invalid input")
	ErrNotFound     = errors.New("memory: not found")
)

// Repository defines how memories are persisted and retrieved.
//
// It keeps memory storage behind an interface so application code can work with
// memories without depending on a specific storage implementation.
type Repository interface {
	// Create stores input as a new Memory.
	//
	// It returns the canonical saved memory so callers receive system-managed
	// fields such as ID and timestamps. It returns ErrInvalidInput when input cannot
	// be accepted.
	Create(ctx context.Context, input CreateInput) (Memory, error)

	// Get retrieves a previously saved Memory by its stable identifier.
	//
	// It gives callers the canonical stored record instead of requiring them to
	// know how memories are located in the backing store. It returns ErrInvalidInput
	// for malformed IDs and ErrNotFound when no memory exists.
	Get(ctx context.Context, id string) (Memory, error)
}
