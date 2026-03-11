package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintHelp_PrintsExpectedSections(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	PrintHelp(&buf)

	s := buf.String()
	if s == "" {
		t.Fatalf("PrintHelp() output is empty")
	}

	for _, want := range []string{
		"memd - local-first, inspectable context memory",
		"Usage:",
		"memd <command>",
		"Commands:",
		"help    Show this help",
		"init    Initialize local memd state and database schema",
		"doctor  Check installation and environment health",
		"ingest  Index a repository into local searchable memory",
		"version Print version information",
		"Exit codes:",
		"0  success",
		"1  error",
		"2  usage",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("expected output to contain %q; got:\n%s", want, s)
		}
	}

	if !strings.HasSuffix(s, "2  usage\n") {
		t.Fatalf("PrintHelp() output should end with %q; got:\n%s", "2  usage\\n", s)
	}
}
