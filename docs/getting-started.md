# Getting Started with memd

## Overview

`memd` is a local-first CLI for managing context memory for AI coding agents.

This guide focuses on first-run setup and day-to-day CLI usage. For the project overview, goals, and developer workflow, see the repository [README](../README.md).

Today, `memd` supports a simple workflow:

1. Initialize local state
2. Verify the local environment
3. Ingest a repository into local searchable memory

All state is stored on your machine. There is no cloud service and no hidden remote state.

## Prerequisites

Before you begin, make sure you have:

- Go `1.26` or newer
- A local checkout of the `memd` repository

Verify your Go version:

```bash
go version
```

## Installation

Clone the repository and move into it:

```bash
git clone https://github.com/sojournerdev/memd.git
cd memd
```

For local development, install the CLI with the Makefile:

```bash
make dev
```

This runs the normal local checks and installs `memd` with build metadata.

To install the CLI directly with Go:

```bash
go install ./cmd/memd
```

Confirm the CLI is available:

```bash
memd help
```

## First Run

### 1. Initialize local state

Run:

```bash
memd init
```

Example output:

```text
OK
state_dir: /Users/you/Library/Application Support/memd
db_path:   /Users/you/Library/Application Support/memd/memd.db
blobs_dir: /Users/you/Library/Application Support/memd/blobs
schema:    ready
```

This creates the local state directory, the SQLite database, and the schema used by `memd`.

### 2. Verify your setup

Run:

```bash
memd doctor
```

Example output:

```text
OK
state_dir:    /Users/you/Library/Application Support/memd
db_path:      /Users/you/Library/Application Support/memd/memd.db
blobs_dir:    /Users/you/Library/Application Support/memd/blobs
journal_mode: wal
sqlite_ver:   3.51.2
foreign_keys: true
busy_timeout: 5000ms
synchronous:  NORMAL (1)
db_writable:  true
```

Use `memd doctor` to confirm that:

- `memd` can resolve its state paths
- the database opens successfully
- SQLite settings are healthy
- the database is writable

### 3. Ingest a repository

To ingest the current directory:

```bash
memd ingest .
```

To ingest a different repository path:

```bash
memd ingest /path/to/repo
```

Example output:

```text
repo:     .
files:    42
chunks:   87
duration: 0.21s
```

If you try to ingest before initialization, `memd` will stop and tell you to run:

```bash
memd init
```

## What `memd ingest` Does

`memd ingest` walks a repository, reads supported text files, splits them into chunks, and stores them in the local SQLite database.

Current behavior:

- Supports common code and text formats such as Go, JavaScript, TypeScript, Python, JSON, YAML, Markdown, and SQL
- Skips common generated and dependency directories such as `.git`, `node_modules`, `dist`, `build`, `target`, and `vendor`
- Skips binary files
- Uses deterministic chunking so repeated ingests are predictable
- Replaces the stored index for the same repository when you ingest it again

## Where Data Is Stored

By default, `memd` follows platform conventions for local application state:

| Platform | Default path                         |
| -------- | ------------------------------------ |
| Linux    | `~/.local/state/memd`                |
| macOS    | `~/Library/Application Support/memd` |
| Windows  | `%AppData%\memd`                     |

The state directory contains:

- `memd.db` for structured SQLite state
- `blobs/` for raw content blobs

## Using a Custom State Directory

Set `MEMD_HOME` to override the default state location:

```bash
MEMD_HOME=/tmp/memd memd doctor
```

This is useful for:

- testing
- isolated local experiments
- temporary state during development

## Troubleshooting

### `memd` is not initialized

If you see an error telling you that `memd` is not initialized, run:

```bash
memd init
```

Then retry your command.

### `memd` is not on your `PATH`

If `memd help` fails after installation, make sure your Go binary directory is on your `PATH`.

For `go install`, this is typically one of:

- `$GOBIN`
- `$GOPATH/bin`

### You want to inspect local state

Run:

```bash
memd doctor
```

This prints the exact paths `memd` is using for local state and the database.

## Current Limitations

`memd` is still in early development.

Current limits include:

- There is no command yet to query, search, or bundle stored data
- Ingest currently uses a fixed list of supported file extensions
- CLI behavior and storage details may still evolve

## Next Steps

After your first successful run, the usual workflow is:

1. Run `memd init` once for a state directory
2. Run `memd doctor` when you want to verify the environment
3. Run `memd ingest <path>` whenever you want to index or refresh a repository
