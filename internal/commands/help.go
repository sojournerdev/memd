package commands

import (
	"fmt"
	"io"
)

func PrintHelp(w io.Writer) {
	fmt.Fprintln(w, `memd - local-first, inspectable context memory for AI coding agents

Usage:
  memd <command>

Commands:
  help			Show this help
  doctor    Check installation and environment health
  version   Print version information
	
Exit codes:
  0  success
  1  error
  2  usage`)
}
