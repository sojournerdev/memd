package commands

import (
	"context"
	"io"
	"time"

	"github.com/sojournerdev/memd/internal/db"
	"github.com/sojournerdev/memd/internal/paths"
	"github.com/sojournerdev/memd/internal/store"
)

const initTimeout = 5 * time.Second

func Init(out, errOut io.Writer) int {
	p, err := paths.Resolve()
	if err != nil {
		tryWritef(errOut, "memd: init: resolve paths: %v\n", err)
		return ExitError
	}

	ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
	defer cancel()

	conn, err := db.Open(ctx, p)
	if err != nil {
		tryWritef(errOut, "memd: init: %v\n", err)
		return ExitError
	}
	// Close is best-effort for this short-lived CLI command.
	defer func() { _ = conn.Close() }()

	if err := store.Migrate(ctx, conn); err != nil {
		tryWritef(errOut, "memd: init: migrate: %v\n", err)
		return ExitError
	}

	if err := writef(
		out,
		"OK\nstate_dir: %s\ndb_path:   %s\nblobs_dir: %s\nschema:    ready\n",
		p.StateDir,
		p.DBPath,
		p.BlobsDir,
	); err != nil {
		return ExitError
	}

	return ExitOK
}
