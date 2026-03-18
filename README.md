# memd

<img src="./assets/readme-gopher.png" width="500" />

Local-first context memory for AI coding agents.

Human-readable · Inspectable · No cloud · No hidden state

## Overview

`memd` is a CLI for managing context memory used by AI coding agents.

It is transparent and predictable by design. Your data lives locally, and you can inspect it at any time.

## Goals

- **Local-first** — state lives on your machine, not in the cloud
- **Inspectable** — simple, human-readable storage
- **Minimal CLI surface** — a small, focused command set

## Requirements
- Go `1.26` or newer

Verify your installed version:

```bash
go version
```

## Installation

Clone the repository:

```bash
git clone https://github.com/sojournerdev/memd.git
cd memd
```

For development, use the Makefile:

```bash
make dev
```

This runs the usual local checks and installs `memd` with build metadata.

Install globally (adds `memd` to your `$PATH` via `$GOBIN` or `$GOPATH/bin`):

```bash
go install ./cmd/memd
memd help
```

## Quick Start

Initialize local state:

```bash
memd init
```

Check that everything is working:

```bash
memd doctor
```

Index the current repository:

```bash
memd ingest .
```

Example output:

```bash
$ memd ingest .
repo:     .
files:    42
chunks:   87
duration: 0.21s
```

## Commands

```bash
memd help       # Show available commands
memd init       # Initialize local memd state and schema
memd ingest .   # Index a repository into local searchable memory
memd version    # Print version information
memd doctor     # Validate local configuration
```

### Verifying your setup

Run `memd doctor` to confirm your installation and see where state is stored:

```bash
$ memd doctor
OK
state_dir:    /home/marcos/.local/state/memd
db_path:      /home/marcos/.local/state/memd/memd.db
blobs_dir:    /home/marcos/.local/state/memd/blobs
journal_mode: wal
sqlite_ver:   3.51.2
foreign_keys: true
busy_timeout: 5000ms
synchronous:  NORMAL (1)
db_writable:  true
```

## Ingest

`memd ingest` walks a repository, reads supported text files, splits them into chunks, and stores them in a local SQLite database.

Right now it:

- supports common code and text file types such as Go, JavaScript, TypeScript, Python, JSON, YAML, Markdown, and SQL
- skips common generated or dependency directories such as `.git`, `node_modules`, `dist`, `build`, `target`, and `vendor`
- skips binary files
- uses deterministic chunking so repeated ingests are predictable
- replaces the stored index for the same repository when you ingest it again

Before using `memd ingest`, run `memd init`.

## Data Storage

`memd` follows platform conventions for local application state:

| Platform | Default path                         |
| -------- | ------------------------------------ |
| Linux    | `~/.local/state/memd`                |
| macOS    | `~/Library/Application Support/memd` |
| Windows  | `%AppData%\memd`                     |

You can override the default state directory with `MEMD_HOME`:

```bash
MEMD_HOME=/tmp/memd memd doctor
```

The storage directory contains:

- `memd.db` — SQLite database for structured state
- `blobs/` — directory for raw content blobs

## Developer Workflow

Use the Makefile for common local tasks:

```bash
make dev
make ci
```

- `make dev` runs the usual local development steps and installs the CLI
- `make ci` runs the stricter CI checks

If you just want a local development install, start with `make dev`.

## Status

**Early development.** The current CLI can initialize local state, run health checks, and ingest repositories into local searchable memory. Core interfaces may still change before a stable release.

## Current Limits

- There is no command yet to query, search, or bundle stored data.
- Ingest currently uses a fixed list of supported file extensions
- Some CLI behavior and storage details may still change

## Contributing

Issues and pull requests are welcome!

## Attribution

The Go Gopher was originally created by Renee French.  
Image sourced from [egonelbre/gophers](https://github.com/egonelbre/gophers) (CC0).  
Go and the Go Gopher are trademarks of Google LLC.

## License

Licensed under the [Apache License 2.0](./LICENSE).
