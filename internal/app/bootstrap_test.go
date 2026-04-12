package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sojournerdev/memd/internal/paths"
)

func TestBootstrapPaths_InitializesApp(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stateDir := filepath.Join(t.TempDir(), "state")
	wantPaths := paths.Paths{
		StateDir: stateDir,
		DBPath:   filepath.Join(stateDir, "memd.db"),
		BlobsDir: filepath.Join(stateDir, "blobs"),
	}
	app, err := BootstrapPaths(ctx, wantPaths)
	if err != nil {
		t.Fatalf("BootstrapPaths() error = %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	if app.Paths != wantPaths {
		t.Fatalf("Paths = %#v, want %#v", app.Paths, wantPaths)
	}
	if app.Memory == nil {
		t.Fatal("Memory = nil, want initialized service")
	}

	for _, path := range []string{app.Paths.StateDir, app.Paths.BlobsDir, app.Paths.DBPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("Stat(%q) error = %v", path, err)
		}
	}

	var count int
	if err := app.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM migrations`).Scan(&count); err != nil {
		t.Fatalf("QueryRowContext(migrations count) error = %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one migration record")
	}
}

func TestBootstrap_ResolvesPaths(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "state")
	t.Setenv(paths.EnvHome, stateDir)
	t.Setenv(paths.EnvXDGStateHome, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app, err := Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	if app.Paths.StateDir != stateDir {
		t.Fatalf("StateDir = %q, want %q", app.Paths.StateDir, stateDir)
	}
	if app.Memory == nil {
		t.Fatal("Memory = nil, want initialized service")
	}
}

func TestBootstrapPaths_InvalidPaths(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := BootstrapPaths(ctx, paths.Paths{})
	if err == nil {
		t.Fatal("BootstrapPaths() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "app: open db:") {
		t.Fatalf("BootstrapPaths() error = %q, want prefix %q", err.Error(), "app: open db:")
	}
}

func TestAppClose_NilSafe(t *testing.T) {
	t.Parallel()

	if err := (*App)(nil).Close(); err != nil {
		t.Fatalf("(*App)(nil).Close() error = %v", err)
	}

	if err := (&App{}).Close(); err != nil {
		t.Fatalf("(&App{}).Close() error = %v", err)
	}
}
