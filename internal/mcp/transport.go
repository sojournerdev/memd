package mcp

import (
	"encoding/json"
	"log/slog"
	"slices"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

type memoryResponse struct {
	ID         string         `json:"id"`
	ProjectKey string         `json:"project_key"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	Content    string         `json:"content"`
	Tags       []string       `json:"tags,omitempty"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

func toMemoryResponse(m memory.Memory) memoryResponse {
	return memoryResponse{
		ID:         m.ID,
		ProjectKey: m.ProjectKey,
		Title:      m.Title,
		Summary:    m.Summary,
		Content:    m.Content,
		Tags:       slices.Clone(m.Tags),
		Metadata:   decodeMetadata(m.Metadata),
		CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:  m.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func encodeMetadata(metadata map[string]any) (json.RawMessage, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func decodeMetadata(metadata json.RawMessage) map[string]any {
	if len(metadata) == 0 {
		return map[string]any{}
	}

	var decoded map[string]any
	if err := json.Unmarshal(metadata, &decoded); err != nil {
		slog.Warn("mcp: failed to decode memory metadata", "err", err)
		return map[string]any{}
	}
	if decoded == nil {
		return map[string]any{}
	}
	return decoded
}
