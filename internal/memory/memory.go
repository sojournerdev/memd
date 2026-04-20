package memory

import (
	"encoding/json"
	"time"
)

// Memory is the application model for a persisted memory artifact.
type Memory struct {
	// ID is assigned when the memory is persisted.
	ID string

	ProjectKey string
	Title      string

	// Summary is the short retrieval-oriented description used for search and recall.
	Summary string

	// Content stores the fuller captured context.
	Content string

	Tags []string

	// Metadata stores optional structured data associated with the memory.
	Metadata json.RawMessage

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateInput is the caller-supplied data required to create a memory.
type CreateInput struct {
	ProjectKey string
	Title      string
	Summary    string
	Content    string
	Tags       []string

	// Metadata stores optional structured data supplied at creation time.
	Metadata json.RawMessage
}
