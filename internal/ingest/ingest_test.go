package ingest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sojournerdev/memd/internal/store"

	_ "modernc.org/sqlite"
)

func TestRun_ValidatesInputs(t *testing.T) {
	t.Parallel()

	if _, err := Run(context.Background(), nil, "."); err == nil {
		t.Fatalf("Run(nil db) error = nil, want error")
	}

	db := openMigratedDB(t)
	if _, err := Run(context.Background(), db, ""); err == nil {
		t.Fatalf("Run(empty repo path) error = nil, want error")
	}
}

func TestStableChunkID_IncludesRepoRoot(t *testing.T) {
	t.Parallel()

	id1 := stableChunkID("/repo/one", "a/b/c.go", 2, "hello")
	id2 := stableChunkID("/repo/one", "a/b/c.go", 2, "hello")
	id3 := stableChunkID("/repo/two", "a/b/c.go", 2, "hello")

	if id1 != id2 {
		t.Fatalf("equal chunk inputs should produce same id")
	}
	if id1 == id3 {
		t.Fatalf("different repo root should produce different id")
	}
}

func TestRun_IngestsDeterministically(t *testing.T) {
	t.Parallel()

	db := openMigratedDB(t)
	repo := t.TempDir()

	mustWriteIngestFile(t, filepath.Join(repo, "a.go"), "package main\n\nfunc main() {}\n")
	mustWriteIngestFile(t, filepath.Join(repo, "README.md"), "hello ingest\n")
	mustWriteIngestFile(t, filepath.Join(repo, "skip.png"), "not indexed\n")
	mustWriteIngestFile(t, filepath.Join(repo, "node_modules", "x.go"), "package x\n")

	res1, err := Run(context.Background(), db, repo)
	if err != nil {
		t.Fatalf("Run() first error = %v", err)
	}
	if res1.Files != 2 {
		t.Fatalf("first files = %d, want %d", res1.Files, 2)
	}
	if res1.Chunks == 0 {
		t.Fatalf("first chunks = 0, want > 0")
	}

	res2, err := Run(context.Background(), db, repo)
	if err != nil {
		t.Fatalf("Run() second error = %v", err)
	}
	if res2 != res1 {
		t.Fatalf("second result = %#v, want %#v", res2, res1)
	}

	repoRoot, err := filepath.Abs(repo)
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}

	assertRepoHasChunks(t, db, repoRoot)
	assertRepoDoesNotIndexFile(t, db, repoRoot, "node_modules/x.go")
	assertRepoDoesNotIndexFile(t, db, repoRoot, "skip.png")

	if !repoContainsChunkText(t, db, repoRoot, "hello ingest") {
		t.Fatalf("repo %q does not contain expected chunk text", repoRoot)
	}
}

func TestRun_IngestsTwoReposWithSamePathsWithoutCollision(t *testing.T) {
	t.Parallel()

	db := openMigratedDB(t)

	repo1 := t.TempDir()
	repo2 := t.TempDir()

	mustWriteIngestFile(t, filepath.Join(repo1, "README.md"), "same content\n")
	mustWriteIngestFile(t, filepath.Join(repo2, "README.md"), "same content\n")

	if _, err := Run(context.Background(), db, repo1); err != nil {
		t.Fatalf("Run(repo1) error = %v", err)
	}
	if _, err := Run(context.Background(), db, repo2); err != nil {
		t.Fatalf("Run(repo2) error = %v", err)
	}

	repoRoot1, err := filepath.Abs(repo1)
	if err != nil {
		t.Fatalf("Abs(repo1) error = %v", err)
	}
	repoRoot2, err := filepath.Abs(repo2)
	if err != nil {
		t.Fatalf("Abs(repo2) error = %v", err)
	}

	assertRepoHasChunks(t, db, repoRoot1)
	assertRepoHasChunks(t, db, repoRoot2)
}

func openMigratedDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	return db
}

func mustWriteIngestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertRepoHasChunks(t *testing.T, db *sql.DB, repoRoot string) {
	t.Helper()

	var chunkCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM repo_chunks WHERE repo_root = ?`, repoRoot).Scan(&chunkCount); err != nil {
		t.Fatalf("count repo_chunks for %q error = %v", repoRoot, err)
	}
	if chunkCount == 0 {
		t.Fatalf("repo_chunks count for %q = 0, want > 0", repoRoot)
	}
}

func assertRepoDoesNotIndexFile(t *testing.T, db *sql.DB, repoRoot, filePath string) {
	t.Helper()

	var fileCount int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM repo_files WHERE repo_root = ? AND file_path = ?`,
		repoRoot,
		filePath,
	).Scan(&fileCount); err != nil {
		t.Fatalf("count repo_files for %q/%q error = %v", repoRoot, filePath, err)
	}
	if fileCount != 0 {
		t.Fatalf("repo_files count for %q/%q = %d, want 0", repoRoot, filePath, fileCount)
	}
}

func repoContainsChunkText(t *testing.T, db *sql.DB, repoRoot, want string) bool {
	t.Helper()

	var matchCount int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM repo_chunks WHERE repo_root = ? AND content LIKE ?`,
		repoRoot,
		"%"+want+"%",
	).Scan(&matchCount); err != nil {
		t.Fatalf("query repo_chunks text for %q error = %v", repoRoot, err)
	}
	return matchCount > 0
}
