package mcp

import (
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

// memoryResponse is the MCP-facing representation of a Memory.
//
// It keeps transport formatting, such as JSON field names and timestamp
// strings, separate from the domain model.
type memoryResponse struct {
	ID         string `json:"id"`
	ProjectKey string `json:"project_key"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type searchMemoriesResponse struct {
	Memories []memoryResponse `json:"memories"`
}

func toMemoryResponse(m memory.Memory) memoryResponse {
	return memoryResponse{
		ID:         m.ID,
		ProjectKey: m.ProjectKey,
		Title:      m.Title,
		Summary:    m.Summary,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:  m.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toMemoryResponses(memories []memory.Memory) []memoryResponse {
	out := make([]memoryResponse, 0, len(memories))
	for _, m := range memories {
		out = append(out, toMemoryResponse(m))
	}
	return out
}
