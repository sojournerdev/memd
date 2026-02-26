package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_OK_Smoke(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	var out, errOut bytes.Buffer
	got := Doctor(&out, &errOut)

	if got != ExitOK {
		t.Fatalf("Doctor() = %d, want %d; errOut=%q", got, ExitOK, errOut.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("errOut = %q, want empty", errOut.String())
	}

	outStr := out.String()
	kv, other := parseDoctorOutput(t, outStr)

	// Success output should contain a single OK line and key/value diagnostics.
	if count(t, other, "OK") != 1 {
		t.Fatalf("expected exactly one OK line; got:\n%s", outStr)
	}
	for _, line := range other {
		if line != "OK" {
			t.Fatalf("unexpected non-kv line %q; got:\n%s", line, outStr)
		}
	}

	for _, k := range []string{
		"state_dir",
		"db_path",
		"blobs_dir",
		"journal_mode",
		"sqlite_ver",
		"foreign_keys",
		"busy_timeout",
		"synchronous",
		"db_writable",
	} {
		if _, ok := kv[k]; !ok {
			t.Fatalf("missing %q in output; got:\n%s", k, outStr)
		}
	}

	clean := filepath.Clean(state)
	if kv["state_dir"] != clean {
		t.Fatalf("state_dir = %q, want %q", kv["state_dir"], clean)
	}
	if kv["db_path"] != filepath.Join(clean, "memd.db") {
		t.Fatalf("db_path = %q, want %q", kv["db_path"], filepath.Join(clean, "memd.db"))
	}
	if kv["blobs_dir"] != filepath.Join(clean, "blobs") {
		t.Fatalf("blobs_dir = %q, want %q", kv["blobs_dir"], filepath.Join(clean, "blobs"))
	}

	if !strings.HasSuffix(kv["busy_timeout"], "ms") {
		t.Fatalf("busy_timeout = %q, want suffix %q", kv["busy_timeout"], "ms")
	}
	if kv["synchronous"] == "" {
		t.Fatalf("synchronous is empty; want non-empty")
	}
}

func TestDoctor_Error_WhenMEMD_HOMEIsFile(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(stateFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("MEMD_HOME", stateFile)

	var out, errOut bytes.Buffer
	got := Doctor(&out, &errOut)

	if got != ExitError {
		t.Fatalf("Doctor() = %d, want %d", got, ExitError)
	}
	if out.Len() != 0 {
		t.Fatalf("out = %q, want empty", out.String())
	}
	if errOut.Len() == 0 {
		t.Fatalf("errOut is empty; want an error message")
	}
	if !strings.Contains(errOut.String(), "memd: doctor:") {
		t.Fatalf("errOut = %q; want to contain %q", errOut.String(), "memd: doctor:")
	}
}

func TestSynchronousLabel(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "OFF"},
		{1, "NORMAL"},
		{2, "FULL"},
		{3, "EXTRA"},
		{42, "UNKNOWN(42)"},
	}

	for _, tt := range tests {
		if got := synchronousLabel(tt.in); got != tt.want {
			t.Fatalf("synchronousLabel(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func parseDoctorOutput(t *testing.T, s string) (map[string]string, []string) {
	t.Helper()

	lines := strings.Split(s, "\n")
	kv := make(map[string]string)
	var other []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		i := strings.IndexByte(line, ':')
		if i <= 0 {
			other = append(other, line)
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])

		if _, exists := kv[key]; exists {
			t.Fatalf("duplicate key %q in output; got:\n%s", key, s)
		}
		kv[key] = val
	}

	return kv, other
}

func count(t *testing.T, lines []string, want string) int {
	t.Helper()

	n := 0
	for _, line := range lines {
		if line == want {
			n++
		}
	}
	return n
}
