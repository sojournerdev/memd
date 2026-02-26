package commands

import (
	"bytes"
	"testing"
)

func TestVersion_PrintsExpectedFormat(t *testing.T) {
	// Mutates package globals; do not run in parallel.

	oldVersion := VersionString
	oldCommit := CommitHash
	oldDate := BuildDate
	t.Cleanup(func() {
		VersionString = oldVersion
		CommitHash = oldCommit
		BuildDate = oldDate
	})

	VersionString = "1.2.3"
	CommitHash = "abc123"
	BuildDate = "2024-01-01"

	var buf bytes.Buffer
	got := Version(&buf)

	if got != ExitOK {
		t.Fatalf("Version() = %d, want %d", got, ExitOK)
	}

	want := "memd 1.2.3 (commit=abc123, built=2024-01-01)\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestVersion_Defaults_AreNonEmpty(t *testing.T) {
	t.Parallel()

	// This test doesn't mutate globals; it just ensures the output is well-formed.
	var buf bytes.Buffer
	got := Version(&buf)

	if got != ExitOK {
		t.Fatalf("Version() = %d, want %d", got, ExitOK)
	}
	if buf.Len() == 0 {
		t.Fatalf("Version() output is empty")
	}
	if buf.Bytes()[buf.Len()-1] != '\n' {
		t.Fatalf("Version() output must end with newline; got %q", buf.String())
	}
}
