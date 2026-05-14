package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sojournerdev/memd/internal/memory"
)

func TestSQLiteRepository_Create_ReturnsCanonicalMemory(t *testing.T) {
	t.Parallel()

	repo := newTestSQLiteRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := repo.Create(ctx, memory.CreateInput{
		ProjectKey: "project-a",
		Title:      "Bootstrap plan",
		Summary:    "Draft summary refined for search.",
		Content:    "Notes about MCP and memory capture.",
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
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("timestamps = (%v, %v), want non-zero", created.CreatedAt, created.UpdatedAt)
	}
	if !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("CreatedAt = %v, UpdatedAt = %v, want equal on create", created.CreatedAt, created.UpdatedAt)
	}
}

func TestSQLiteRepository_Search_FindsCreatedMemory(t *testing.T) {
	t.Parallel()

	repo := newTestSQLiteRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := repo.Create(ctx, memory.CreateInput{
		ProjectKey: "project-a",
		Title:      "Bootstrap plan",
		Summary:    "Draft summary refined for search.",
		Content:    "Notes about MCP and memory capture.",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	results, err := repo.Search(ctx, memory.SearchInput{
		ProjectKey: "project-a",
		Query:      "bootstrap",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Search() result count = %d, want 1", len(results))
	}
	if results[0] != created {
		t.Fatalf("Search()[0] = %#v, want %#v", results[0], created)
	}
}

func TestSQLiteRepository_New_ReturnsErrorForNilDB(t *testing.T) {
	t.Parallel()

	if _, err := NewSQLiteRepository(nil); err == nil {
		t.Fatal("NewSQLiteRepository(nil) error = nil, want error")
	} else if !strings.Contains(err.Error(), "nil db") {
		t.Fatalf("NewSQLiteRepository(nil) error = %q, want substring %q", err.Error(), "nil db")
	}
}

func newTestSQLiteRepository(t *testing.T) *SQLiteRepository {
	t.Helper()

	db := migratedTestDB(t)
	repo, err := NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	return repo
}
