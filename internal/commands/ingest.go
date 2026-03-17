package commands

import (
	"context"
	"io"
	"time"

	"github.com/sojournerdev/memd/internal/db"
	"github.com/sojournerdev/memd/internal/ingest"
	"github.com/sojournerdev/memd/internal/paths"
	"github.com/sojournerdev/memd/internal/store"
)

const (
	ingestTimeout     = 10 * time.Minute
	defaultIngestPath = "."
)

func Ingest(args []string, out, errOut io.Writer) int {
	if len(args) > 1 {
		tryWrite(errOut, "memd: ingest: too many arguments\nUsage: memd ingest [path]\n")
		return ExitUsage
	}

	repoPath := defaultIngestPath
	if len(args) == 1 {
		repoPath = args[0]
	}

	var err error
	p, err := paths.Resolve()
	if err != nil {
		tryWritef(errOut, "memd: ingest: resolve paths: %v\n", err)
		return ExitError
	}

	ctx, cancel := context.WithTimeout(context.Background(), ingestTimeout)
	defer cancel()

	conn, err := db.Open(ctx, p)
	if err != nil {
		tryWritef(errOut, "memd: ingest: %v\n", err)
		return ExitError
	}
	defer func() { _ = conn.Close() }()

	ready, err := store.IngestSchemaReady(ctx, conn)
	if err != nil {
		tryWritef(errOut, "memd: ingest: verify database readiness: %v\n", err)
		return ExitError
	}
	if !ready {
		tryWrite(errOut, "memd is not initialized.\nRun: memd init\n")
		return ExitError
	}

	start := time.Now()
	result, err := ingest.Run(ctx, conn, repoPath)
	if err != nil {
		tryWritef(errOut, "memd: ingest: %v\n", err)
		return ExitError
	}

	if err := writef(
		out,
		"repo:     %s\nfiles:    %d\nchunks:   %d\nduration: %.2fs\n",
		repoPath,
		result.Files,
		result.Chunks,
		time.Since(start).Seconds(),
	); err != nil {
		return ExitError
	}

	return ExitOK
}
