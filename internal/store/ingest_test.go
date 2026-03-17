package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func TestIngestSchemaReady(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	ready, err := IngestSchemaReady(ctx, db)
	if err != nil {
		t.Fatalf("IngestSchemaReady() before migrate error = %v", err)
	}
	if ready {
		t.Fatalf("IngestSchemaReady() before migrate = true, want false")
	}

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	ready, err = IngestSchemaReady(ctx, db)
	if err != nil {
		t.Fatalf("IngestSchemaReady() after migrate error = %v", err)
	}
	if !ready {
		t.Fatalf("IngestSchemaReady() after migrate = false, want true")
	}
}

func TestIngestSchemaReady_FailsWhenRequiredIndexMissing(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	if _, err := db.ExecContext(ctx, `DROP INDEX idx_repo_chunks_repo_root`); err != nil {
		t.Fatalf("DROP INDEX error = %v", err)
	}

	ready, err := IngestSchemaReady(ctx, db)
	if err != nil {
		t.Fatalf("IngestSchemaReady() error = %v", err)
	}
	if ready {
		t.Fatalf("IngestSchemaReady() = true, want false when required index is missing")
	}
}

func TestReplaceRepoIndex_ValidatesInputs(t *testing.T) {
	t.Parallel()

	err := ReplaceRepoIndex(context.Background(), nil, "/repo", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "nil db") {
		t.Fatalf("ReplaceRepoIndex(nil db) error = %v, want nil db error", err)
	}

	db := openTestDB(t)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	err = ReplaceRepoIndex(context.Background(), db, "", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "empty repo root") {
		t.Fatalf("ReplaceRepoIndex(empty repo root) error = %v, want empty repo root error", err)
	}
}

func TestReplaceRepoIndex_ReplacesOnlyTargetRepoAndUpdatesFTS(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	repoA := "/repo-a"
	repoB := "/repo-b"

	initialFiles := []RepoFileRecord{
		{
			RepoRoot:      repoA,
			FilePath:      "README.md",
			ContentSHA256: "file-a-1",
			ByteSize:      5,
			ChunkCount:    1,
		},
		{
			RepoRoot:      repoB,
			FilePath:      "README.md",
			ContentSHA256: "file-b-1",
			ByteSize:      4,
			ChunkCount:    1,
		},
	}
	initialChunks := []RepoChunkRecord{
		{
			ChunkID:       "chunk-a-1",
			RepoRoot:      repoA,
			FilePath:      "README.md",
			ChunkIndex:    0,
			Content:       "alpha",
			ContentSHA256: "chunk-a-sha-1",
		},
		{
			ChunkID:       "chunk-b-1",
			RepoRoot:      repoB,
			FilePath:      "README.md",
			ChunkIndex:    0,
			Content:       "bravo",
			ContentSHA256: "chunk-b-sha-1",
		},
	}

	if err := ReplaceRepoIndex(ctx, db, repoA, initialFiles[:1], initialChunks[:1]); err != nil {
		t.Fatalf("ReplaceRepoIndex(repoA initial) error = %v", err)
	}
	if err := ReplaceRepoIndex(ctx, db, repoB, initialFiles[1:], initialChunks[1:]); err != nil {
		t.Fatalf("ReplaceRepoIndex(repoB initial) error = %v", err)
	}

	replacedFiles := []RepoFileRecord{
		{
			RepoRoot:      repoA,
			FilePath:      "README.md",
			ContentSHA256: "file-a-2",
			ByteSize:      5,
			ChunkCount:    1,
		},
	}
	replacedChunks := []RepoChunkRecord{
		{
			ChunkID:       "chunk-a-2",
			RepoRoot:      repoA,
			FilePath:      "README.md",
			ChunkIndex:    0,
			Content:       "apple",
			ContentSHA256: "chunk-a-sha-2",
		},
	}

	if err := ReplaceRepoIndex(ctx, db, repoA, replacedFiles, replacedChunks); err != nil {
		t.Fatalf("ReplaceRepoIndex(repoA replace) error = %v", err)
	}

	assertFTSHits(t, db, "apple", 1)
	assertFTSHits(t, db, "alpha", 0)
	assertFTSHits(t, db, "bravo", 1)

	assertRepoChunkCount(t, db, repoA, 1)
	assertRepoChunkCount(t, db, repoB, 1)
}

func assertFTSHits(t *testing.T, db *sql.DB, term string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM repo_chunks_fts
		WHERE repo_chunks_fts MATCH ?
	`, term).Scan(&got); err != nil {
		t.Fatalf("FTS query for %q error = %v", term, err)
	}
	if got != want {
		t.Fatalf("FTS hits for %q = %d, want %d", term, got, want)
	}
}

func assertRepoChunkCount(t *testing.T, db *sql.DB, repoRoot string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM repo_chunks WHERE repo_root = ?`, repoRoot).Scan(&got); err != nil {
		t.Fatalf("count repo_chunks for %q error = %v", repoRoot, err)
	}
	if got != want {
		t.Fatalf("repo_chunks count for %q = %d, want %d", repoRoot, got, want)
	}
}
