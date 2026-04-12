package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

func TestSQLiteRepository_CreateAndGet_RoundTripCanonicalMemory(t *testing.T) {
	t.Parallel()

	repo, db := newTestSQLiteRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := repo.Create(ctx, memory.CreateInput{
		ProjectKey: "  project-a  ",
		Title:      "  Bootstrap plan  ",
		Summary:    "  Draft summary refined for search.  ",
		Content:    "  Notes about MCP and memory capture.  ",
		Tags:       []string{"mcp", "architecture"},
		Metadata:   json.RawMessage(` { "source": "chat", "kind": "context-save" } `),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID == "" {
		t.Fatal("Create() returned empty ID")
	}
	if created.ProjectKey != "project-a" {
		t.Fatalf("ProjectKey = %q, want %q", created.ProjectKey, "project-a")
	}
	if created.Title != "Bootstrap plan" {
		t.Fatalf("Title = %q, want %q", created.Title, "Bootstrap plan")
	}
	if created.Summary != "Draft summary refined for search." {
		t.Fatalf("Summary = %q, want %q", created.Summary, "Draft summary refined for search.")
	}
	if created.Content != "Notes about MCP and memory capture." {
		t.Fatalf("Content = %q, want %q", created.Content, "Notes about MCP and memory capture.")
	}
	if got, want := string(created.Metadata), `{"source":"chat","kind":"context-save"}`; got != want {
		t.Fatalf("Metadata = %q, want %q", got, want)
	}
	if len(created.Tags) != 2 || created.Tags[0] != "architecture" || created.Tags[1] != "mcp" {
		t.Fatalf("Tags = %#v, want %#v", created.Tags, []string{"architecture", "mcp"})
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("timestamps = (%v, %v), want non-zero", created.CreatedAt, created.UpdatedAt)
	}
	if !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("CreatedAt = %v, UpdatedAt = %v, want equal on create", created.CreatedAt, created.UpdatedAt)
	}

	stored := countRows(t, db, `SELECT COUNT(*) FROM memories WHERE memory_id = ?`, created.ID)
	if stored != 1 {
		t.Fatalf("stored memories row count = %d, want 1", stored)
	}

	var searchTags string
	if err := db.QueryRowContext(ctx, `SELECT tags FROM memory_search WHERE memory_id = ?`, created.ID).Scan(&searchTags); err != nil {
		t.Fatalf("QueryRowContext(memory_search tags) error = %v", err)
	}
	if searchTags != "architecture mcp" {
		t.Fatalf("memory_search tags = %q, want %q", searchTags, "architecture mcp")
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got, created) {
		t.Fatalf("Get() = %#v, want %#v", got, created)
	}
}

func TestSQLiteRepository_Create_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	repo, _ := newTestSQLiteRepository(t)

	for _, tc := range []struct {
		name  string
		input memory.CreateInput
	}{
		{
			name: "blank project key",
			input: memory.CreateInput{
				ProjectKey: " ",
				Title:      "Title",
				Summary:    "Summary",
				Content:    "Content",
			},
		},
		{
			name: "blank title",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      " ",
				Summary:    "Summary",
				Content:    "Content",
			},
		},
		{
			name: "blank summary",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    " ",
				Content:    "Content",
			},
		},
		{
			name: "blank content",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    "Summary",
				Content:    " ",
			},
		},
		{
			name: "blank tag",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    "Summary",
				Content:    "Content",
				Tags:       []string{"tag-a", " "},
			},
		},
		{
			name: "duplicate tags",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    "Summary",
				Content:    "Content",
				Tags:       []string{"tag-a", "tag-a"},
			},
		},
		{
			name: "metadata is not object",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    "Summary",
				Content:    "Content",
				Metadata:   json.RawMessage(`[]`),
			},
		},
		{
			name: "metadata is malformed json",
			input: memory.CreateInput{
				ProjectKey: "project-a",
				Title:      "Title",
				Summary:    "Summary",
				Content:    "Content",
				Metadata:   json.RawMessage(`{"source":`),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := repo.Create(context.Background(), tc.input)
			if !errors.Is(err, memory.ErrInvalidInput) {
				t.Fatalf("Create() error = %v, want %v", err, memory.ErrInvalidInput)
			}
		})
	}
}

func TestSQLiteRepository_Create_DefaultsMetadataObject(t *testing.T) {
	t.Parallel()

	repo, _ := newTestSQLiteRepository(t)

	got, err := repo.Create(context.Background(), memory.CreateInput{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if string(got.Metadata) != "{}" {
		t.Fatalf("Metadata = %q, want %q", got.Metadata, "{}")
	}
}

func TestSQLiteRepository_Create_DoesNotRetainCallerMetadata(t *testing.T) {
	t.Parallel()

	repo, _ := newTestSQLiteRepository(t)

	input := memory.CreateInput{
		ProjectKey: "project-a",
		Title:      "Title",
		Summary:    "Summary",
		Content:    "Content",
		Metadata:   json.RawMessage(`{"source":"chat"}`),
	}

	got, err := repo.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	input.Metadata[0] = '['

	if string(got.Metadata) != `{"source":"chat"}` {
		t.Fatalf("Metadata after caller mutation = %q, want %q", got.Metadata, `{"source":"chat"}`)
	}
}

func TestSQLiteRepository_Get_ReturnsErrorsForInvalidOrMissingID(t *testing.T) {
	t.Parallel()

	repo, _ := newTestSQLiteRepository(t)

	for _, tc := range []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "blank id", id: " ", wantErr: memory.ErrInvalidInput},
		{name: "missing id", id: "missing", wantErr: memory.ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := repo.Get(context.Background(), tc.id)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Get(%q) error = %v, want %v", tc.id, err, tc.wantErr)
			}
		})
	}
}

func TestSQLiteRepository_Methods_ReturnErrorForNilReceiver(t *testing.T) {
	t.Parallel()

	var repo *SQLiteRepository

	_, err := repo.Create(context.Background(), memory.CreateInput{})
	if !errors.Is(err, errNilDB) {
		t.Fatalf("nil Create() error = %v, want %v", err, errNilDB)
	}

	_, err = repo.Get(context.Background(), "mem_123")
	if !errors.Is(err, errNilDB) {
		t.Fatalf("nil Get() error = %v, want %v", err, errNilDB)
	}
}

func TestSQLiteRepository_New_ReturnsErrorForNilDB(t *testing.T) {
	t.Parallel()

	_, err := NewSQLiteRepository(nil)
	if err == nil {
		t.Fatal("NewSQLiteRepository(nil) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "nil db") {
		t.Fatalf("NewSQLiteRepository(nil) error = %q, want substring %q", err.Error(), "nil db")
	}
}

func newTestSQLiteRepository(t *testing.T) (*SQLiteRepository, *sql.DB) {
	t.Helper()

	db := migratedTestDB(t)
	repo, err := NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	return repo, db
}
