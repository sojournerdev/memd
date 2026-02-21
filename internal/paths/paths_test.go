package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_EnvPrecedenceAndShape(t *testing.T) {
	tests := []struct {
		name     string
		memdHome string
		xdgHome  string
		want     func(memdHome, xdgHome string) string
	}{
		{
			name:     "MEMD_HOME wins",
			memdHome: "SET_BY_TEST",
			xdgHome:  "IGNORED",
			want: func(memdHome, _ string) string {
				return filepath.Clean(memdHome)
			},
		},
		{
			name:     "XDG_STATE_HOME used when MEMD_HOME empty",
			memdHome: "",
			xdgHome:  "SET_BY_TEST",
			want: func(_ string, xdgHome string) string {
				return filepath.Clean(filepath.Join(xdgHome, AppName))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Use temp dirs + a dirty path to validate cleaning without OS-specific assumptions.
			memd := tt.memdHome
			xdg := tt.xdgHome

			if memd == "SET_BY_TEST" {
				memd = filepath.Join(t.TempDir(), "memd-home", ".", "x", "..")
			}
			if xdg == "SET_BY_TEST" || xdg == "IGNORED" {
				xdg = t.TempDir()
			}

			t.Setenv(EnvHome, memd)
			t.Setenv("XDG_STATE_HOME", xdg)

			p, err := Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			wantState := tt.want(memd, xdg)
			assertEqual(t, "StateDir", p.StateDir, wantState)
			assertEqual(t, "DBPath", p.DBPath, filepath.Join(wantState, "memd.db"))
			assertEqual(t, "BlobsDir", p.BlobsDir, filepath.Join(wantState, "blobs"))
		})
	}
}

func TestResolve_DefaultStateDir_InvariantsOnly(t *testing.T) {
	// OS-dependent; only assert invariants to avoid brittle expectations.
	t.Setenv(EnvHome, "")
	t.Setenv("XDG_STATE_HOME", "")

	p, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if p.StateDir == "" {
		t.Fatalf("StateDir is empty; want non-empty")
	}
	if !strings.HasSuffix(filepath.Clean(p.StateDir), AppName) {
		t.Fatalf("StateDir = %q; want suffix %q", p.StateDir, AppName)
	}

	state := filepath.Clean(p.StateDir)
	assertEqual(t, "DBPath", p.DBPath, filepath.Join(state, "memd.db"))
	assertEqual(t, "BlobsDir", p.BlobsDir, filepath.Join(state, "blobs"))
}

func TestEnsure_CreatesDirectories(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state")
	p := Paths{
		StateDir: state,
		DBPath:   filepath.Join(state, "memd.db"),
		BlobsDir: filepath.Join(state, "blobs"),
	}

	if err := p.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	assertDir(t, p.StateDir)
	assertDir(t, p.BlobsDir)
}

func TestEnsure_EmptyStateDir(t *testing.T) {
	err := (Paths{StateDir: ""}).Ensure()
	assertErrContains(t, err, "paths: empty StateDir")
}

func TestValidateReadWrite_HappyPath(t *testing.T) {
	if err := (Paths{StateDir: t.TempDir()}).ValidateReadWrite(); err != nil {
		t.Fatalf("ValidateReadWrite() error = %v", err)
	}
}

func TestValidateReadWrite_EmptyStateDir(t *testing.T) {
	err := (Paths{StateDir: ""}).ValidateReadWrite()
	assertErrContains(t, err, "paths: empty StateDir")
}

func TestValidateReadWrite_StateDirMissing(t *testing.T) {
	err := (Paths{StateDir: filepath.Join(t.TempDir(), "missing")}).ValidateReadWrite()
	if err == nil {
		t.Fatalf("ValidateReadWrite() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "paths: stat state dir:") {
		t.Fatalf("error = %q; want to contain %q", err.Error(), "paths: stat state dir:")
	}
}

func TestValidateReadWrite_StateDirIsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := (Paths{StateDir: path}).ValidateReadWrite()
	if err == nil {
		t.Fatalf("ValidateReadWrite() expected error, got nil")
	}
	assertErrContains(t, err, "state dir is not a directory")
}

func assertDir(t *testing.T, path string) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if !fi.IsDir() {
		t.Fatalf("%q exists but is not a directory", path)
	}
}

func assertEqual[T comparable](t *testing.T, field string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %v; want %v", field, got, want)
	}
}

func assertErrContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("error = %q; want to contain %q", err.Error(), substr)
	}
}
