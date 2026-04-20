# memd

<img src="./assets/readme-gopher.png" width="500" />

Local-first memory for AI coding agents, exposed through the Model Context
Protocol (MCP).

`memd` lets an AI assistant save and retrieve durable project context without
shipping that context to a hosted memory service.

## Overview

`memd` is a personal tool that came from an OCD-ish annoyance: I wanted to clean
up old chats in the VS Code Codex extension without feeling like I was losing
useful context. It lets me save the parts worth keeping before deleting the noisy
session, so I can pull that context into another chat later.

## Status

Early development. The current server exposes a small MCP surface for creating
and retrieving memories. The capture workflow is still evolving.

## Principles

- **Local-first**: memories are stored on your machine.
- **Agent-friendly**: MCP tools give AI clients a direct integration point.
- **Inspectable**: data is stored in SQLite.
- **Small core**: persistence, application logic, and MCP transport stay
  separate.

## How It Works

`memd` runs as an MCP stdio server. An MCP-capable client, such as VS Code or
Codex, starts the server and discovers its tools.

Current tools:

- `create_memory`: persist a finalized memory artifact.
- `get_memory`: retrieve a memory by its stable ID.

The MCP boundary automatically adds small provenance tags and metadata so saved
memories can be traced back to how they were captured.

## Requirements

- Go `1.26` or newer
- An MCP-capable client for interactive use

## Quick Start

Build and run the MCP server:

```sh
make run
```

For MCP clients, configure the server command as:

```sh
go run ./cmd/memd
```

Example VS Code workspace config:

```json
{
  "servers": {
    "memd": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "./cmd/memd"],
      "cwd": "${workspaceFolder}"
    }
  }
}
```

Example Codex local config:

```toml
[mcp_servers.memd]
command = "go"
args = ["run", "./cmd/memd"]
cwd = "/path/to/memd"
```

## Example Tool Calls

Create a memory:

```json
{
  "project_key": "memd",
  "title": "MCP smoke test",
  "summary": "Testing that memd can persist and retrieve memories through MCP.",
  "content": "This memory verifies the create/get round trip through the MCP server."
}
```

Retrieve it:

```json
{
  "id": "mem_..."
}
```

## Local State

State is stored locally in SQLite. `memd` resolves its state directory from
`MEMD_HOME`, `XDG_STATE_HOME`, or the OS default.

## Development

```sh
make build
make run
make ci
```

- `make build` runs formatting, vet, tests, and builds `./.bin/memd`.
- `make run` builds and starts the MCP server.
- `make ci` runs the stricter local CI path.

## Architecture

The code is split by responsibility:

- `cmd/memd`: process entrypoint and build metadata.
- `internal/app`: application bootstrap and dependency wiring.
- `internal/mcp`: MCP transport, tool registration, and request/response shapes.
- `internal/memory`: memory domain types, service, and repository contract.
- `internal/store`: SQLite implementation and migrations.
- `internal/db`: SQLite opening and PRAGMA configuration.
- `internal/paths`: local state path resolution.

## Current Limits

- Search/list tools are not implemented yet.
- `create_memory` expects a finalized memory artifact.
- The higher-level “save this chat context” drafting workflow is planned but not
  implemented yet.
- Metadata is intentionally small and currently focused on capture provenance.

## Roadmap

- Add search backed by the existing FTS table.
- Add a higher-level context-capture workflow.
- Add confirmation/refinement flow for generated title and summary.
- Improve observability around MCP calls.

## Attribution

The Go Gopher was originally created by Renee French.
Image sourced from [egonelbre/gophers](https://github.com/egonelbre/gophers)
(CC0).
Go and the Go Gopher are trademarks of Google LLC.

## License

Licensed under the [Apache License 2.0](./LICENSE).
