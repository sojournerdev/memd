package commands

import (
	"fmt"
	"io"
)

var (
	VersionString = "dev"
	CommitHash    = "unknown"
	BuildDate     = "unknown"
)

func Version(out io.Writer) int {
	fmt.Fprintf(out, "memd %s (commit=%s, built=%s)\n", VersionString, CommitHash, BuildDate)
	return ExitOK
}
