package ingest

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWalk_DeterministicAndFiltered(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustWriteFile(t, filepath.Join(repo, "z.go"), "package z\n")
	mustWriteFile(t, filepath.Join(repo, "a.md"), "# a\n")
	mustWriteFile(t, filepath.Join(repo, "sub", "b.sql"), "select 1;\n")
	mustWriteFile(t, filepath.Join(repo, "sub", "c.txt"), "notes\n")
	mustWriteFile(t, filepath.Join(repo, "sub", "ignored.json"), "{}\n")
	mustWriteFile(t, filepath.Join(repo, "node_modules", "x.go"), "package x\n")
	mustWriteFile(t, filepath.Join(repo, ".git", "config"), "[core]\n")

	got, err := Walk(context.Background(), repo)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	want := []string{"a.md", "sub/b.sql", "sub/c.txt", "sub/ignored.json", "z.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Walk() = %#v, want %#v", got, want)
	}
}

func TestWalk_CanceledContext(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	mustWriteFile(t, filepath.Join(repo, "a.go"), "package main\n")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Walk(ctx, repo)
	if err == nil {
		t.Fatalf("Walk() error = nil, want context cancellation error")
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
