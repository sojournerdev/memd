package memory

import (
	"encoding/json"
	"time"
)

// Memory represents chat context captured for reuse in future conversations.
//
// It preserves durable details worth carrying forward so old chats can be
// cleaned up while later conversations still receive relevant context.
type Memory struct {
	ID         string
	ProjectKey string
	Title      string
	Summary    string
	Content    string
	Tags       []string
	Metadata   json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreateInput contains the client-provided data needed to create a Memory.
//
// It captures the proposed memory content before validation, persistence, and
// assignment of system-managed fields such as ID and timestamps.
type CreateInput struct {
	ProjectKey string
	Title      string
	Summary    string
	Content    string
	Tags       []string
	Metadata   json.RawMessage
}
