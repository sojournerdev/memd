package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sojournerdev/memd/internal/memory"
)

const serverName = "memd"

// Options configures metadata advertised by the MCP server.
type Options struct {
	Version string
}

// Server exposes memd application services over MCP.
//
// It keeps transport concerns in this package while memory rules remain in the
// memory package.
type Server struct {
	srv *mcp.Server
}

// New returns an MCP server wired to the memory service.
//
// It registers the tools clients use to create and retrieve memories.
func New(memoryService *memory.Service, opts Options) *Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: opts.Version,
	}, nil)

	addCreateMemoryTool(srv, memoryService)
	addGetMemoryTool(srv, memoryService)

	return &Server{srv: srv}
}

// RunStdio serves MCP over stdin/stdout until the client disconnects or ctx is
// canceled.
func (s *Server) RunStdio(ctx context.Context) error {
	return s.srv.Run(ctx, &mcp.StdioTransport{})
}
