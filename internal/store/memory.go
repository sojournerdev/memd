package store

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

const emptyMetadataJSON = `{}`

var errNilDB = errors.New("store: nil db")

// SQLiteRepository implements memory.Repository using SQLite.
type SQLiteRepository struct {
	db *sql.DB
}

var _ memory.Repository = (*SQLiteRepository)(nil)

// NewSQLiteRepository returns a SQLite-backed memory repository.
func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	if db == nil {
		return nil, errNilDB
	}
	return &SQLiteRepository{db: db}, nil
}

// Create validates and persists a memory in a single transaction.
func (r *SQLiteRepository) Create(ctx context.Context, input memory.CreateInput) (memory.Memory, error) {
	if r == nil || r.db == nil {
		return memory.Memory{}, errNilDB
	}
	if ctx == nil {
		ctx = context.Background()
	}

	record, err := newMemoryRecord(input)
	if err != nil {
		return memory.Memory{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("store: begin create memory tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO memories (
			memory_id, project_key, title, summary, content, metadata_json, created_at_ns, updated_at_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.ProjectKey,
		record.Title,
		record.Summary,
		record.Content,
		string(record.Metadata),
		record.CreatedAt.UnixNano(),
		record.UpdatedAt.UnixNano(),
	); err != nil {
		return memory.Memory{}, fmt.Errorf("store: insert memory: %w", err)
	}

	for _, tag := range record.Tags {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO memory_tags (memory_id, tag) VALUES (?, ?)`,
			record.ID,
			tag,
		); err != nil {
			return memory.Memory{}, fmt.Errorf("store: insert memory tag %q: %w", tag, err)
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO memory_search (memory_id, project_key, title, summary, content, tags) VALUES (?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.ProjectKey,
		record.Title,
		record.Summary,
		record.Content,
		strings.Join(record.Tags, " "),
	); err != nil {
		return memory.Memory{}, fmt.Errorf("store: insert memory search row: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return memory.Memory{}, fmt.Errorf("store: commit create memory tx: %w", err)
	}

	return record, nil
}

// Get loads a persisted memory and its tags by identifier.
func (r *SQLiteRepository) Get(ctx context.Context, id string) (memory.Memory, error) {
	if r == nil || r.db == nil {
		return memory.Memory{}, errNilDB
	}
	if ctx == nil {
		ctx = context.Background()
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return memory.Memory{}, memory.ErrInvalidInput
	}

	var (
		record                   memory.Memory
		metadata                 string
		createdAtNS, updatedAtNS int64
	)
	err := r.db.QueryRowContext(
		ctx,
		`SELECT memory_id, project_key, title, summary, content, metadata_json, created_at_ns, updated_at_ns
		 FROM memories
		 WHERE memory_id = ?`,
		id,
	).Scan(
		&record.ID,
		&record.ProjectKey,
		&record.Title,
		&record.Summary,
		&record.Content,
		&metadata,
		&createdAtNS,
		&updatedAtNS,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return memory.Memory{}, memory.ErrNotFound
		}
		return memory.Memory{}, fmt.Errorf("store: get memory %q: %w", id, err)
	}

	tags, err := loadTags(ctx, r.db, id)
	if err != nil {
		return memory.Memory{}, err
	}

	record.Metadata = json.RawMessage(metadata)
	record.Tags = tags
	record.CreatedAt = time.Unix(0, createdAtNS).UTC()
	record.UpdatedAt = time.Unix(0, updatedAtNS).UTC()
	return record, nil
}

func loadTags(ctx context.Context, db *sql.DB, id string) ([]string, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT tag FROM memory_tags WHERE memory_id = ? ORDER BY tag`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("store: query tags for memory %q: %w", id, err)
	}
	defer func() { _ = rows.Close() }()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("store: scan tag for memory %q: %w", id, err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate tags for memory %q: %w", id, err)
	}
	return tags, nil
}

func newMemoryRecord(input memory.CreateInput) (memory.Memory, error) {
	projectKey := strings.TrimSpace(input.ProjectKey)
	title := strings.TrimSpace(input.Title)
	summary := strings.TrimSpace(input.Summary)
	content := strings.TrimSpace(input.Content)

	switch {
	case projectKey == "":
		return memory.Memory{}, memory.ErrInvalidInput
	case title == "":
		return memory.Memory{}, memory.ErrInvalidInput
	case summary == "":
		return memory.Memory{}, memory.ErrInvalidInput
	case content == "":
		return memory.Memory{}, memory.ErrInvalidInput
	}

	tags, err := normalizeTags(input.Tags)
	if err != nil {
		return memory.Memory{}, err
	}

	metadata, err := normalizeMetadata(input.Metadata)
	if err != nil {
		return memory.Memory{}, err
	}

	id, err := newMemoryID()
	if err != nil {
		return memory.Memory{}, fmt.Errorf("store: generate memory id: %w", err)
	}

	now := time.Now().UTC()
	return memory.Memory{
		ID:         id,
		ProjectKey: projectKey,
		Title:      title,
		Summary:    summary,
		Content:    content,
		Tags:       tags,
		Metadata:   metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, memory.ErrInvalidInput
		}
		if _, ok := seen[tag]; ok {
			return nil, memory.ErrInvalidInput
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	slices.Sort(out)
	return out, nil
}

func normalizeMetadata(metadata json.RawMessage) (json.RawMessage, error) {
	if len(metadata) == 0 {
		return json.RawMessage(emptyMetadataJSON), nil
	}

	trimmed := bytes.TrimSpace(metadata)
	if len(trimmed) == 0 {
		return json.RawMessage(emptyMetadataJSON), nil
	}

	var decoded any
	if err := json.Unmarshal(trimmed, &decoded); err != nil {
		return nil, memory.ErrInvalidInput
	}
	if _, ok := decoded.(map[string]any); !ok {
		return nil, memory.ErrInvalidInput
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, trimmed); err != nil {
		return nil, memory.ErrInvalidInput
	}
	return bytes.Clone(compact.Bytes()), nil
}

func newMemoryID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "mem_" + hex.EncodeToString(b[:]), nil
}
