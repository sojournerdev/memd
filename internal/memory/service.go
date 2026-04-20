package memory

import (
	"context"
	"errors"
)

var errNilRepository = errors.New("memory: nil repository")

// Service provides application-level memory operations.
//
// It is the home for memory use cases and business rules. Persistence is
// delegated through Repository so callers do not depend on storage details.
type Service struct {
	repo Repository
}

// NewService returns a Service backed by repo.
func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errNilRepository
	}
	return &Service{repo: repo}, nil
}

// Create creates a memory through the configured repository.
func (s *Service) Create(ctx context.Context, input CreateInput) (Memory, error) {
	if s == nil || s.repo == nil {
		return Memory{}, errNilRepository
	}
	return s.repo.Create(ctx, input)
}

// Get retrieves a memory by its stable identifier.
func (s *Service) Get(ctx context.Context, id string) (Memory, error) {
	if s == nil || s.repo == nil {
		return Memory{}, errNilRepository
	}
	return s.repo.Get(ctx, id)
}
