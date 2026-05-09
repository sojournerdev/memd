package memory

import (
	"context"
	"errors"
)

var errNilRepository = errors.New("memory: nil repository")

// Service coordinates memory use cases for the application.
//
// It keeps memory business rules in one place while delegating persistence to
// Repository so callers do not depend on storage details.
type Service struct {
	repo Repository
}

// NewService returns a Service that uses repo for persistence.
func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errNilRepository
	}
	return &Service{repo: repo}, nil
}

// Create captures input as a memory for future recall.
func (s *Service) Create(ctx context.Context, input CreateInput) (Memory, error) {
	if s == nil || s.repo == nil {
		return Memory{}, errNilRepository
	}
	return s.repo.Create(ctx, input)
}

// Get returns a saved memory for reuse in application workflows.
func (s *Service) Get(ctx context.Context, id string) (Memory, error) {
	if s == nil || s.repo == nil {
		return Memory{}, errNilRepository
	}
	return s.repo.Get(ctx, id)
}
