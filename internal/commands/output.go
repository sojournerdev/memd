package commands

import (
	"fmt"
	"io"
)

// tryWrite is for best-effort plain text output (help/usage/error printing).
func tryWrite(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}

// writef is for command success output where write failures should fail command execution.
func writef(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

// tryWritef is for best-effort output (help/usage/error printing).
func tryWritef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
