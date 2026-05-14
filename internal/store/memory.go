package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

// SQLiteRepository stores memories in SQLite.
//
// It implements memory.Repository so the application can use SQLite without
// depending on SQLite details outside the store package.
type SQLiteRepository struct {
	db *sql.DB
}

var _ memory.Repository = (*SQLiteRepository)(nil)

// NewSQLiteRepository returns a memory repository backed by db.
func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("store: nil db")
	}
	return &SQLiteRepository{db: db}, nil
}

// Create validates input and stores a complete memory record.
//
// It writes the memory and search data together so callers receive a saved
// Memory that is ready for retrieval.
func (r *SQLiteRepository) Create(ctx context.Context, input memory.CreateInput) (memory.Memory, error) {
	id, err := newMemoryID()
	if err != nil {
		return memory.Memory{}, fmt.Errorf("store: generate memory id: %w", err)
	}

	now := time.Now().UTC()
	record := memory.Memory{
		ID:         id,
		ProjectKey: input.ProjectKey,
		Title:      input.Title,
		Summary:    input.Summary,
		Content:    input.Content,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("store: begin create memory tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO memories (
			memory_id, project_key, title, summary, content, created_at_ns, updated_at_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.ProjectKey,
		record.Title,
		record.Summary,
		record.Content,
		record.CreatedAt.UnixNano(),
		record.UpdatedAt.UnixNano(),
	); err != nil {
		return memory.Memory{}, fmt.Errorf("store: insert memory: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO memory_search (
			memory_id, project_key, title, summary, content
		) VALUES (?, ?, ?, ?, ?)`,
		record.ID,
		record.ProjectKey,
		record.Title,
		record.Summary,
		record.Content,
	); err != nil {
		return memory.Memory{}, fmt.Errorf("store: insert memory search row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return memory.Memory{}, fmt.Errorf("store: commit create memory tx: %w", err)
	}

	return record, nil
}

// Search finds saved memories by matching the persisted search index.
//
// It keeps recall backed by SQLite FTS so callers can rediscover context by
// topic instead of needing to know a memory's generated ID.
func (r *SQLiteRepository) Search(ctx context.Context, input memory.SearchInput) ([]memory.Memory, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT m.memory_id, m.project_key, m.title, m.summary, m.content,
		        m.created_at_ns, m.updated_at_ns
		 FROM memory_search
		 JOIN memories m ON m.memory_id = memory_search.memory_id
		 WHERE memory_search MATCH ?
		   AND memory_search.project_key = ?
		 ORDER BY bm25(memory_search) ASC, m.updated_at_ns DESC, m.memory_id
		 LIMIT ?`,
		input.Query,
		input.ProjectKey,
		input.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("store: search memories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []memory.Memory
	for rows.Next() {
		var (
			record                   memory.Memory
			createdAtNS, updatedAtNS int64
		)
		if err := rows.Scan(
			&record.ID,
			&record.ProjectKey,
			&record.Title,
			&record.Summary,
			&record.Content,
			&createdAtNS,
			&updatedAtNS,
		); err != nil {
			return nil, fmt.Errorf("store: scan memory search result: %w", err)
		}

		record.CreatedAt = time.Unix(0, createdAtNS).UTC()
		record.UpdatedAt = time.Unix(0, updatedAtNS).UTC()
		results = append(results, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate memory search results: %w", err)
	}

	return results, nil
}

func newMemoryID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "mem_" + hex.EncodeToString(b[:]), nil
}
