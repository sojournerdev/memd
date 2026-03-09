package commands

import "io"

// helpText is kept as a constant so help output is deterministic.
const helpText = `memd - local-first, inspectable context memory for AI coding agents

Usage:
  memd <command>

Commands:
  help    Show this help
  init    Initialize local memd state and database schema
  doctor  Check installation and environment health
  version Print version information

Exit codes:
  0  success
  1  error
  2  usage
`

func PrintHelp(w io.Writer) {
	tryWrite(w, helpText)
}
