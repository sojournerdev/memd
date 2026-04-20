package mcp

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sojournerdev/memd/internal/memory"
)

const serverName = "memd"

// Options configures process metadata for the MCP server.
type Options struct {
	Version string
}

// Server exposes memd's application services over the MCP transport.
//
// It is intentionally transport-focused while memory rules remain in the
// memory package.
type Server struct {
	server *mcp.Server
}

// New returns an MCP server exposing the memory operations needed for manual
// testing and early client integration.
func New(memories *memory.Service, opts Options) (*Server, error) {
	if memories == nil {
		return nil, errors.New("mcp: nil memory service")
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: opts.Version,
	}, nil)

	addCreateMemoryTool(server, memories)
	addGetMemoryTool(server, memories)

	return &Server{server: server}, nil
}

// RunStdio serves MCP over stdin/stdout until the client disconnects or ctx is
// canceled.
func (s *Server) RunStdio(ctx context.Context) error {
	if s == nil || s.server == nil {
		return errors.New("mcp: nil server")
	}
	return s.server.Run(ctx, &mcp.StdioTransport{})
}
