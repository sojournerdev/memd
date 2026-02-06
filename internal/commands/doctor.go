package commands

import (
	"fmt"
	"io"

	"github.com/sojournerdev/memd/internal/paths"
)

func Doctor(out, errOut io.Writer) int {
	p, err := paths.Resolve()
	if err != nil {
		fmt.Fprintf(errOut, "memd: doctor: resolve paths: %v\n", err)
		return ExitError
	}
	if err := p.Ensure(); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: ensure dirs: %v\n", err)
		return ExitError
	}
	if err := p.ValidateReadWrite(); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: validate read/write: %v\n", err)
		return ExitError
	}

	fmt.Fprintln(out, "OK")
	fmt.Fprintf(out, "state_dir: %s\n", p.StateDir)
	fmt.Fprintf(out, "db_path:   %s\n", p.DBPath)
	fmt.Fprintf(out, "blobs_dir: %s\n", p.BlobsDir)
	return ExitOK
}
