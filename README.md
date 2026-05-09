# memd

<p align="center">
  <img
    src="./assets/banner.png"
    alt="memd - clean chats, keep what matters"
    width="760"
  />
</p>

## Overview

`memd` lets you save useful context from AI chats and reuse it later, so you can
clean up old conversations without losing the details you still want.

`memd` comes from my OCD of cleaning up conversations when I use Codex. As I
write this, my understanding is that you can't delete chats inside the VS Code
extension and have to go to the backend of Codex, where it stores your chat
sessions on disk. So I thought it would be nice to delete them inside VS Code
using tool calls. The added benefit is saving helpful info that I want to pull
in later.

## Goals

- Save useful context from AI chats so it can be reused later.
- Help keep chat history clean without losing important details.
- Store memories locally in a format that is easy to inspect.
- Expose an MCP surface that AI clients can use directly.

## Non-Goals

- Replace a full notes app or knowledge base.
- Store entire chat transcripts by default.
- Sync memories across machines.
- Decide automatically what should be remembered without user review.

## How It Works

`memd` runs as an MCP stdio server. An MCP-capable client, such as VS Code or
Codex, starts the server and discovers its tools.

Current tools:

- `create_memory`: save useful context for later.
- `get_memory`: retrieve a saved memory by its stable ID.

The MCP boundary automatically adds small provenance tags and metadata so saved
memories can be traced back to how they were captured.

## Requirements

- Go `1.26.2` or newer
- An MCP-capable client for interactive use

## Quick Start

Build and run the MCP server:

```sh
make
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
make
make build
make run
make ci
```

- `make` builds and starts the MCP server.
- `make build` runs formatting, vet, tests, and builds `./.bin/memd`.
- `make run` builds and starts the MCP server.
- `make ci` runs the stricter local CI path.

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

## License

Licensed under [MIT](./LICENSE).
