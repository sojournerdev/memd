package mcp

import (
	"context"
	"errors"

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
	server *mcp.Server
}

// New returns an MCP server wired to the memory service.
//
// It registers the tools clients use to create and retrieve memories.
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
