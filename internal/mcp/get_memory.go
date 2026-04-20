package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sojournerdev/memd/internal/memory"
)

const getMemoryToolName = "get_memory"

type getMemoryRequest struct {
	ID string `json:"id" jsonschema:"stable memory identifier"`
}

func addGetMemoryTool(server *sdkmcp.Server, memories *memory.Service) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        getMemoryToolName,
		Description: "Load a persisted memory by ID.",
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, req getMemoryRequest) (*sdkmcp.CallToolResult, memoryResponse, error) {
		got, err := memories.Get(ctx, req.ID)
		if err != nil {
			return nil, memoryResponse{}, err
		}
		return nil, toMemoryResponse(got), nil
	})
}
