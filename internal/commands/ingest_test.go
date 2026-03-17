package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIngest_Error_WhenNotInitialized(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	repo := t.TempDir()
	if err := writeFile(filepath.Join(repo, "a.go"), "package main\n"); err != nil {
		t.Fatalf("writeFile() error = %v", err)
	}

	var out, errOut bytes.Buffer
	got := Ingest([]string{repo}, &out, &errOut)

	if got != ExitError {
		t.Fatalf("Ingest() = %d, want %d", got, ExitError)
	}
	if out.Len() != 0 {
		t.Fatalf("out = %q, want empty", out.String())
	}
	if !strings.Contains(errOut.String(), "memd is not initialized.") {
		t.Fatalf("errOut missing not initialized guidance; got:\n%s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Run: memd init") {
		t.Fatalf("errOut missing init guidance; got:\n%s", errOut.String())
	}
}

func TestIngest_UsageError_WhenTooManyArgs(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	got := Ingest([]string{"repo-a", "repo-b"}, &out, &errOut)

	if got != ExitUsage {
		t.Fatalf("Ingest() = %d, want %d", got, ExitUsage)
	}
	if out.Len() != 0 {
		t.Fatalf("out = %q, want empty", out.String())
	}
	if !strings.Contains(errOut.String(), "too many arguments") {
		t.Fatalf("errOut missing usage error; got:\n%s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Usage: memd ingest [path]") {
		t.Fatalf("errOut missing usage text; got:\n%s", errOut.String())
	}
}

func TestIngest_OK_SummaryAndIdempotent(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	var initOut, initErr bytes.Buffer
	if got := Init(&initOut, &initErr); got != ExitOK {
		t.Fatalf("Init() = %d, want %d; errOut=%q", got, ExitOK, initErr.String())
	}

	repo := t.TempDir()
	if err := writeFile(filepath.Join(repo, "a.go"), "package main\n\nfunc main() {}\n"); err != nil {
		t.Fatalf("writeFile(a.go) error = %v", err)
	}
	if err := writeFile(filepath.Join(repo, "README.md"), "hello memd ingest\n"); err != nil {
		t.Fatalf("writeFile(README.md) error = %v", err)
	}
	if err := writeFile(filepath.Join(repo, "skip.bin"), string([]byte{0x00, 0x01, 0x02})); err != nil {
		t.Fatalf("writeFile(skip.bin) error = %v", err)
	}
	if err := writeFile(filepath.Join(repo, "node_modules", "a.txt"), "skip me\n"); err != nil {
		t.Fatalf("writeFile(node_modules/a.txt) error = %v", err)
	}

	var out1, errOut1 bytes.Buffer
	got1 := Ingest([]string{repo}, &out1, &errOut1)
	if got1 != ExitOK {
		t.Fatalf("Ingest() first run = %d, want %d; errOut=%q", got1, ExitOK, errOut1.String())
	}
	if errOut1.Len() != 0 {
		t.Fatalf("first run errOut = %q, want empty", errOut1.String())
	}
	for _, want := range []string{"repo:", "files:", "chunks:", "duration:"} {
		if !strings.Contains(out1.String(), want) {
			t.Fatalf("first run out missing %q; got:\n%s", want, out1.String())
		}
	}

	var out2, errOut2 bytes.Buffer
	got2 := Ingest([]string{repo}, &out2, &errOut2)
	if got2 != ExitOK {
		t.Fatalf("Ingest() second run = %d, want %d; errOut=%q", got2, ExitOK, errOut2.String())
	}
	if errOut2.Len() != 0 {
		t.Fatalf("second run errOut = %q, want empty", errOut2.String())
	}
	for _, want := range []string{"repo:", "files:", "chunks:", "duration:"} {
		if !strings.Contains(out2.String(), want) {
			t.Fatalf("second run out missing %q; got:\n%s", want, out2.String())
		}
	}
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
