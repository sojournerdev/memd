package mcp

import (
	"context"
	"maps"
	"slices"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sojournerdev/memd/internal/memory"
)

const (
	createMemoryToolName = "create_memory"
	getMemoryToolName    = "get_memory"
)

func addTool[In, Out any](
	server *sdkmcp.Server,
	name string,
	description string,
	handler func(context.Context, *sdkmcp.CallToolRequest, In) (*sdkmcp.CallToolResult, Out, error),
) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        name,
		Description: description,
	}, handler)
}

type createMemoryRequest struct {
	ProjectKey string         `json:"project_key" jsonschema:"project or workspace key for the memory"`
	Title      string         `json:"title" jsonschema:"short human-readable title"`
	Summary    string         `json:"summary" jsonschema:"short retrieval-oriented summary"`
	Content    string         `json:"content" jsonschema:"full memory content to persist"`
	Tags       []string       `json:"tags,omitempty" jsonschema:"optional tags for retrieval and filtering"`
	Metadata   map[string]any `json:"metadata,omitempty" jsonschema:"optional metadata object"`
}

func addCreateMemoryTool(server *sdkmcp.Server, memoryService *memory.Service) {
	addTool(server, createMemoryToolName, "Persist a memory artifact in memd.",
		func(ctx context.Context, _ *sdkmcp.CallToolRequest, req createMemoryRequest) (*sdkmcp.CallToolResult, memoryResponse, error) {
			createInput, err := req.toCreateInput()
			if err != nil {
				return nil, memoryResponse{}, err
			}

			created, err := memoryService.Create(ctx, createInput)
			if err != nil {
				return nil, memoryResponse{}, err
			}
			return nil, toMemoryResponse(created), nil
		})
}

func (req createMemoryRequest) toCreateInput() (memory.CreateInput, error) {
	tags := mergeCreateTags(req.Tags)
	metadata, err := encodeMetadata(mergeCreateMetadata(req.Metadata))
	if err != nil {
		return memory.CreateInput{}, err
	}

	return memory.CreateInput{
		ProjectKey: req.ProjectKey,
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Tags:       tags,
		Metadata:   metadata,
	}, nil
}

type getMemoryRequest struct {
	ID string `json:"id" jsonschema:"stable memory identifier"`
}

func addGetMemoryTool(server *sdkmcp.Server, memoryService *memory.Service) {
	addTool(server, getMemoryToolName, "Load a persisted memory by ID.",
		func(ctx context.Context, _ *sdkmcp.CallToolRequest, req getMemoryRequest) (*sdkmcp.CallToolResult, memoryResponse, error) {
			got, err := memoryService.Get(ctx, req.ID)
			if err != nil {
				return nil, memoryResponse{}, err
			}
			return nil, toMemoryResponse(got), nil
		})
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
