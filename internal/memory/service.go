package memory

import (
	"context"
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

	// Search finds saved memories relevant to input.
	//
	// It gives callers a topic-based recall path so they do not need to know
	// system-generated memory IDs.
	Search(ctx context.Context, input SearchInput) ([]Memory, error)
}

// Service coordinates memory use cases for the application.
//
// It keeps memory business rules in one place while delegating persistence to
// Repository so callers do not depend on storage details.
type Service struct {
	repo Repository
}

// NewService returns a Service that uses repo for persistence.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Memory, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Search(ctx context.Context, input SearchInput) ([]Memory, error) {
	return s.repo.Search(ctx, input)
}
