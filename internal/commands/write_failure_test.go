package commands

import (
	"bytes"
	"io"
	"testing"
)

type errWriter struct {
	err error
}

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestVersion_ReturnsExitError_WhenOutWriteFails(t *testing.T) {
	got := Version(errWriter{err: io.ErrClosedPipe})
	if got != ExitError {
		t.Fatalf("Version() = %d, want %d", got, ExitError)
	}
}

func TestInit_ReturnsExitError_WhenOutWriteFails(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	var errOut bytes.Buffer
	got := Init(errWriter{err: io.ErrClosedPipe}, &errOut)
	if got != ExitError {
		t.Fatalf("Init() = %d, want %d; errOut=%q", got, ExitError, errOut.String())
	}
}

func TestDoctor_ReturnsExitError_WhenOutWriteFails(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	var errOut bytes.Buffer
	got := Doctor(errWriter{err: io.ErrClosedPipe}, &errOut)
	if got != ExitError {
		t.Fatalf("Doctor() = %d, want %d; errOut=%q", got, ExitError, errOut.String())
	}
}

func TestPrintHelp_DoesNotPanic_WhenWriteFails(t *testing.T) {
	PrintHelp(errWriter{err: io.ErrClosedPipe})
}
