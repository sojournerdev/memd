package commands

import (
	"context"
	"fmt"
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
		fmt.Fprintf(errOut, "memd: init: resolve paths: %v\n", err)
		return ExitError
	}

	ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
	defer cancel()

	conn, err := db.Open(ctx, p)
	if err != nil {
		fmt.Fprintf(errOut, "memd: init: %v\n", err)
		return ExitError
	}
	defer conn.Close()

	if err := store.Migrate(ctx, conn); err != nil {
		fmt.Fprintf(errOut, "memd: init: migrate: %v\n", err)
		return ExitError
	}

	fmt.Fprintln(out, "OK")
	fmt.Fprintf(out, "state_dir: %s\n", p.StateDir)
	fmt.Fprintf(out, "db_path:   %s\n", p.DBPath)
	fmt.Fprintf(out, "blobs_dir: %s\n", p.BlobsDir)
	fmt.Fprintln(out, "schema:    ready")

	return ExitOK
}
