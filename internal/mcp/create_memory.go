package mcp

import (
	"context"
	"maps"
	"slices"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sojournerdev/memd/internal/memory"
)

const createMemoryToolName = "create_memory"

type createMemoryRequest struct {
	ProjectKey string         `json:"project_key" jsonschema:"project or workspace key for the memory"`
	Title      string         `json:"title" jsonschema:"short human-readable title"`
	Summary    string         `json:"summary" jsonschema:"short retrieval-oriented summary"`
	Content    string         `json:"content" jsonschema:"full memory content to persist"`
	Tags       []string       `json:"tags,omitempty" jsonschema:"optional tags for retrieval and filtering"`
	Metadata   map[string]any `json:"metadata,omitempty" jsonschema:"optional metadata object"`
}

func addCreateMemoryTool(server *sdkmcp.Server, memories *memory.Service) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        createMemoryToolName,
		Description: "Persist a memory artifact in memd.",
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, req createMemoryRequest) (*sdkmcp.CallToolResult, memoryResponse, error) {
		createInput, err := req.toCreateInput()
		if err != nil {
			return nil, memoryResponse{}, err
		}

		created, err := memories.Create(ctx, createInput)
		if err != nil {
			return nil, memoryResponse{}, err
		}
		return nil, toMemoryResponse(created), nil
	})
}

func (req createMemoryRequest) toCreateInput() (memory.CreateInput, error) {
	metadata, err := encodeMetadata(mergeCreateMetadata(req.Metadata))
	if err != nil {
		return memory.CreateInput{}, err
	}

	return memory.CreateInput{
		ProjectKey: req.ProjectKey,
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Tags:       mergeCreateTags(req.Tags),
		Metadata:   metadata,
	}, nil
}

func mergeCreateMetadata(metadata map[string]any) map[string]any {
	out := make(map[string]any, len(metadata)+3)
	maps.Copy(out, metadata)
	out["capture_client"] = "codex"
	out["capture_source"] = "mcp"
	out["capture_tool"] = createMemoryToolName
	return out
}

func mergeCreateTags(tags []string) []string {
	out := make([]string, 0, len(tags)+2)
	out = append(out, "codex", "mcp")
	out = append(out, tags...)
	slices.Sort(out)
	return slices.Compact(out)
}
