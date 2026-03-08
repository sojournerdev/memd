package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_Init_RoutesAndSucceeds(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	var out, errOut bytes.Buffer
	got := Run([]string{"memd", "init"}, &out, &errOut)

	if got != ExitOK {
		t.Fatalf("Run(init) = %d, want %d; errOut=%q", got, ExitOK, errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("Run(init) errOut = %q, want empty", errOut.String())
	}
	if !strings.Contains(out.String(), "schema:    ready") {
		t.Fatalf("Run(init) out missing schema marker; got:\n%s", out.String())
	}
}

func TestRun_Init_RoutesAndPropagatesFailure(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(stateFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("MEMD_HOME", stateFile)

	var out, errOut bytes.Buffer
	got := Run([]string{"memd", "init"}, &out, &errOut)

	if got != ExitError {
		t.Fatalf("Run(init) = %d, want %d", got, ExitError)
	}
	if out.Len() != 0 {
		t.Fatalf("Run(init) out = %q, want empty", out.String())
	}
	if errOut.Len() == 0 {
		t.Fatal("Run(init) errOut is empty; want error output")
	}
	if !strings.Contains(errOut.String(), "memd: init:") {
		t.Fatalf("Run(init) errOut = %q, want to contain %q", errOut.String(), "memd: init:")
	}
}
