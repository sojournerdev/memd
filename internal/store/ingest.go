package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type RepoFileRecord struct {
	RepoRoot      string
	FilePath      string
	ContentSHA256 string
	ByteSize      int
	ChunkCount    int
}

type RepoChunkRecord struct {
	ChunkID       string
	RepoRoot      string
	FilePath      string
	ChunkIndex    int
	Content       string
	ContentSHA256 string
}

type schemaObject struct {
	name string
	typ  string
}

var requiredIngestSchemaObjects = []schemaObject{
	{name: "migrations", typ: "table"},
	{name: "repo_files", typ: "table"},
	{name: "repo_chunks", typ: "table"},
	{name: "repo_chunks_fts", typ: "table"},
	{name: "idx_repo_chunks_repo_file_chunk", typ: "index"},
	{name: "idx_repo_chunks_repo_root", typ: "index"},
}

// IngestSchemaReady reports whether the required ingest schema objects exist.
func IngestSchemaReady(ctx context.Context, db *sql.DB) (bool, error) {
	if db == nil {
		return false, errors.New("store: ingest schema ready: nil db")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for _, obj := range requiredIngestSchemaObjects {
		exists, err := schemaObjectExists(ctx, db, obj.typ, obj.name)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}

	return true, nil
}

// ReplaceRepoIndex replaces all stored files, chunks, and FTS rows for one repo.
func ReplaceRepoIndex(ctx context.Context, db *sql.DB, repoRoot string, files []RepoFileRecord, chunks []RepoChunkRecord) error {
	if db == nil {
		return errors.New("store: replace repo index: nil db")
	}
	if repoRoot == "" {
		return errors.New("store: replace repo index: empty repo root")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: replace repo index: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM repo_chunks_fts
		WHERE chunk_id IN (
			SELECT chunk_id
			FROM repo_chunks
			WHERE repo_root = ?
		)
	`, repoRoot); err != nil {
		return fmt.Errorf("store: replace repo index: clear repo fts rows: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM repo_chunks WHERE repo_root = ?`, repoRoot); err != nil {
		return fmt.Errorf("store: replace repo index: clear repo chunks: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM repo_files WHERE repo_root = ?`, repoRoot); err != nil {
		return fmt.Errorf("store: replace repo index: clear repo files: %w", err)
	}

	fileStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO repo_files(repo_root, file_path, content_sha256, byte_size, chunk_count)
		VALUES(?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("store: replace repo index: prepare file insert: %w", err)
	}
	defer func() { _ = fileStmt.Close() }()

	chunkStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO repo_chunks(chunk_id, repo_root, file_path, chunk_index, content, content_sha256)
		VALUES(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("store: replace repo index: prepare chunk insert: %w", err)
	}
	defer func() { _ = chunkStmt.Close() }()

	ftsStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO repo_chunks_fts(chunk_id, content)
		VALUES(?, ?)
	`)
	if err != nil {
		return fmt.Errorf("store: replace repo index: prepare fts insert: %w", err)
	}
	defer func() { _ = ftsStmt.Close() }()

	for _, rec := range files {
		if _, err := fileStmt.ExecContext(ctx,
			rec.RepoRoot,
			rec.FilePath,
			rec.ContentSHA256,
			rec.ByteSize,
			rec.ChunkCount,
		); err != nil {
			return fmt.Errorf("store: replace repo index: insert file %q: %w", rec.FilePath, err)
		}
	}

	for _, rec := range chunks {
		if _, err := chunkStmt.ExecContext(ctx,
			rec.ChunkID,
			rec.RepoRoot,
			rec.FilePath,
			rec.ChunkIndex,
			rec.Content,
			rec.ContentSHA256,
		); err != nil {
			return fmt.Errorf("store: replace repo index: insert chunk %q[%d]: %w", rec.FilePath, rec.ChunkIndex, err)
		}
		if _, err := ftsStmt.ExecContext(ctx, rec.ChunkID, rec.Content); err != nil {
			return fmt.Errorf("store: replace repo index: insert fts chunk %q[%d]: %w", rec.FilePath, rec.ChunkIndex, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: replace repo index: commit: %w", err)
	}

	return nil
}

func schemaObjectExists(ctx context.Context, db *sql.DB, typ, name string) (bool, error) {
	var n int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?`,
		typ,
		name,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("store: ingest schema ready: check %s %q: %w", typ, name, err)
	}
	return n == 1, nil
}
