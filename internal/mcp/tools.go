package mcp

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sojournerdev/memd/internal/memory"
)

const (
	createMemoryToolName   = "create_memory"
	searchMemoriesToolName = "search_memories"

	defaultSearchLimit = 10
	maxSearchLimit     = 50
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
	ProjectKey string `json:"project_key" jsonschema:"project or workspace key for the memory"`
	Title      string `json:"title" jsonschema:"short human-readable title"`
	Summary    string `json:"summary" jsonschema:"short retrieval-oriented summary"`
	Content    string `json:"content" jsonschema:"full memory content to persist"`
}

func addCreateMemoryTool(server *sdkmcp.Server, memoryService *memory.Service) {
	addTool(server, createMemoryToolName, "Persist a memory artifact in memd.",
		func(ctx context.Context, _ *sdkmcp.CallToolRequest, req createMemoryRequest) (*sdkmcp.CallToolResult, memoryResponse, error) {
			start := time.Now()
			createInput, err := req.toCreateInput()
			if err != nil {
				slog.Error(
					"mcp tool request validation failed",
					"tool", createMemoryToolName,
					"project_key", createInput.ProjectKey,
					"title", createInput.Title,
					"content_bytes", len(createInput.Content),
					"duration_ms", time.Since(start).Milliseconds(),
					"err", err,
				)
				return nil, memoryResponse{}, err
			}

			slog.Info(
				"mcp tool request started",
				"tool", createMemoryToolName,
				"project_key", createInput.ProjectKey,
				"title", createInput.Title,
				"content_bytes", len(createInput.Content),
			)

			created, err := memoryService.Create(ctx, createInput)
			if err != nil {
				slog.Error(
					"mcp tool request failed",
					"tool", createMemoryToolName,
					"project_key", createInput.ProjectKey,
					"title", createInput.Title,
					"content_bytes", len(createInput.Content),
					"duration_ms", time.Since(start).Milliseconds(),
					"err", err,
				)
				return nil, memoryResponse{}, err
			}
			slog.Info(
				"mcp tool request completed",
				"tool", createMemoryToolName,
				"project_key", created.ProjectKey,
				"title", created.Title,
				"memory_id", created.ID,
				"content_bytes", len(created.Content),
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return nil, toMemoryResponse(created), nil
		})
}

func (req createMemoryRequest) toCreateInput() (memory.CreateInput, error) {
	input := memory.CreateInput{
		ProjectKey: strings.TrimSpace(req.ProjectKey),
		Title:      strings.TrimSpace(req.Title),
		Summary:    strings.TrimSpace(req.Summary),
		Content:    strings.TrimSpace(req.Content),
	}
	if input.ProjectKey == "" || input.Title == "" || input.Summary == "" || input.Content == "" {
		return input, errors.New("mcp: invalid create memory request")
	}

	return input, nil
}

type searchMemoriesRequest struct {
	ProjectKey string `json:"project_key" jsonschema:"project or workspace key to search within"`
	Query      string `json:"query" jsonschema:"topic or text to search for"`
	Limit      int    `json:"limit,omitempty" jsonschema:"maximum number of memories to return"`
}

func addSearchMemoriesTool(server *sdkmcp.Server, memoryService *memory.Service) {
	addTool(server, searchMemoriesToolName, "Search saved memories by topic.",
		func(ctx context.Context, _ *sdkmcp.CallToolRequest, req searchMemoriesRequest) (*sdkmcp.CallToolResult, searchMemoriesResponse, error) {
			start := time.Now()
			searchInput, err := req.toSearchInput()
			if err != nil {
				slog.Error(
					"mcp tool request validation failed",
					"tool", searchMemoriesToolName,
					"project_key", searchInput.ProjectKey,
					"query", searchInput.Query,
					"limit", searchInput.Limit,
					"duration_ms", time.Since(start).Milliseconds(),
					"err", err,
				)
				return nil, searchMemoriesResponse{}, err
			}

			slog.Info(
				"mcp tool request started",
				"tool", searchMemoriesToolName,
				"project_key", searchInput.ProjectKey,
				"query", searchInput.Query,
				"limit", searchInput.Limit,
			)

			results, err := memoryService.Search(ctx, searchInput)
			if err != nil {
				slog.Error(
					"mcp tool request failed",
					"tool", searchMemoriesToolName,
					"project_key", searchInput.ProjectKey,
					"query", searchInput.Query,
					"limit", searchInput.Limit,
					"duration_ms", time.Since(start).Milliseconds(),
					"err", err,
				)
				return nil, searchMemoriesResponse{}, err
			}
			slog.Info(
				"mcp tool request completed",
				"tool", searchMemoriesToolName,
				"project_key", searchInput.ProjectKey,
				"query", searchInput.Query,
				"limit", searchInput.Limit,
				"result_count", len(results),
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return nil, searchMemoriesResponse{Memories: toMemoryResponses(results)}, nil
		})
}

func (req searchMemoriesRequest) toSearchInput() (memory.SearchInput, error) {
	input := memory.SearchInput{
		ProjectKey: strings.TrimSpace(req.ProjectKey),
		Query:      strings.TrimSpace(req.Query),
		Limit:      req.Limit,
	}
	if input.ProjectKey == "" || input.Query == "" || input.Limit < 0 {
		return input, errors.New("mcp: invalid search memories request")
	}
	if input.Limit == 0 {
		input.Limit = defaultSearchLimit
	}
	if input.Limit > maxSearchLimit {
		input.Limit = maxSearchLimit
	}

	return input, nil
}
