package commands

import "io"

// These are overridden at build time via -ldflags.
var (
	VersionString = "dev"
	CommitHash    = "unknown"
	BuildDate     = "unknown"
)

func Version(out io.Writer) int {
	if err := writef(out, "memd %s (commit=%s, built=%s)\n", VersionString, CommitHash, BuildDate); err != nil {
		return ExitError
	}
	return ExitOK
}
