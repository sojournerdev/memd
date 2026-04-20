package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrate_CreatesExpectedSchema(t *testing.T) {
	t.Parallel()

	db := migratedTestDB(t)

	for _, object := range []struct {
		objectType string
		name       string
	}{
		{objectType: "table", name: "memories"},
		{objectType: "table", name: "memory_tags"},
		{objectType: "table", name: "migrations"},
		{objectType: "index", name: "idx_memories_project_updated"},
		{objectType: "index", name: "idx_memory_tags_tag"},
		{objectType: "table", name: "memory_search"},
	} {
		if !sqliteObjectExists(t, db, object.objectType, object.name) {
			t.Fatalf("expected %s %q to exist after migration", object.objectType, object.name)
		}
	}

	if got := countRows(t, db, `SELECT COUNT(*) FROM migrations`); got == 0 {
		t.Fatal("expected at least one migration record after migration")
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := migratedTestDB(t)

	firstRunCount := countRows(t, db, `SELECT COUNT(*) FROM migrations`)
	if firstRunCount == 0 {
		t.Fatal("expected first migration run to record at least one row")
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() second run error = %v", err)
	}

	if got := countRows(t, db, `SELECT COUNT(*) FROM migrations`); got != firstRunCount {
		t.Fatalf("migration row count after second run = %d, want %d", got, firstRunCount)
	}
}

func TestMigrate_EnforcesSchemaConstraints(t *testing.T) {
	t.Parallel()

	db := migratedTestDB(t)

	for _, tc := range []struct {
		name string
		args []any
	}{
		{
			name: "blank title",
			args: []any{"mem-1", "project-a", "   ", "summary", "content", `{}`, 1, 1},
		},
		{
			name: "non object metadata_json",
			args: []any{"mem-2", "project-a", "Title", "summary", "content", `[]`, 1, 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.Exec(`
				INSERT INTO memories (
					memory_id, project_key, title, summary, content, metadata_json, created_at_ns, updated_at_ns
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, tc.args...)
			if err == nil {
				t.Fatalf("expected insert to fail for %s", tc.name)
			}
		})
	}
}

func TestMigrate_AllowsValidMemoryInsert(t *testing.T) {
	t.Parallel()

	db := migratedTestDB(t)

	_, err := db.Exec(`
		INSERT INTO memories (
			memory_id, project_key, title, summary, content, metadata_json, created_at_ns, updated_at_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "mem-1", "project-a", "Title", "summary", "content", `{"source":"test"}`, 1, 2)
	if err != nil {
		t.Fatalf("INSERT INTO memories error = %v", err)
	}

	if got := countRows(t, db, `SELECT COUNT(*) FROM memories WHERE memory_id = ?`, "mem-1"); got != 1 {
		t.Fatalf("inserted memories row count = %d, want 1", got)
	}
}

func TestMigrate_MemoryTagsForeignKeyCascade(t *testing.T) {
	t.Parallel()

	db := migratedTestDB(t)

	_, err := db.Exec(`
		INSERT INTO memories (
			memory_id, project_key, title, summary, content, metadata_json, created_at_ns, updated_at_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "mem-1", "project-a", "Title", "summary", "content", `{}`, 1, 2)
	if err != nil {
		t.Fatalf("INSERT INTO memories error = %v", err)
	}

	_, err = db.Exec(`INSERT INTO memory_tags (memory_id, tag) VALUES (?, ?)`, "missing-memory", "orphan")
	if err == nil {
		t.Fatal("expected orphan memory_tags insert to fail")
	}

	_, err = db.Exec(`INSERT INTO memory_tags (memory_id, tag) VALUES (?, ?)`, "mem-1", "tag-a")
	if err != nil {
		t.Fatalf("INSERT INTO memory_tags error = %v", err)
	}

	if got := countRows(t, db, `SELECT COUNT(*) FROM memory_tags WHERE memory_id = ?`, "mem-1"); got != 1 {
		t.Fatalf("memory_tags row count before delete = %d, want 1", got)
	}

	_, err = db.Exec(`DELETE FROM memories WHERE memory_id = ?`, "mem-1")
	if err != nil {
		t.Fatalf("DELETE FROM memories error = %v", err)
	}

	if got := countRows(t, db, `SELECT COUNT(*) FROM memory_tags WHERE memory_id = ?`, "mem-1"); got != 0 {
		t.Fatalf("memory_tags row count after delete = %d, want 0", got)
	}
}

func TestMigrate_FailsOnChecksumMismatch(t *testing.T) {
	t.Parallel()

	db := migratedTestDB(t)

	var version int
	if err := db.QueryRow(`SELECT version FROM migrations ORDER BY version LIMIT 1`).Scan(&version); err != nil {
		t.Fatalf("QueryRow(first migration version) error = %v", err)
	}

	_, err := db.Exec(`UPDATE migrations SET checksum = ? WHERE version = ?`, "not-the-real-checksum", version)
	if err != nil {
		t.Fatalf("UPDATE migrations checksum error = %v", err)
	}

	err = Migrate(context.Background(), db)
	if err == nil {
		t.Fatal("Migrate() error = nil, want checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("Migrate() error = %q, want substring %q", err.Error(), "checksum mismatch")
	}
}

func TestMigrate_NilDB(t *testing.T) {
	t.Parallel()

	err := Migrate(context.Background(), nil)
	if err == nil {
		t.Fatal("Migrate() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "nil db") {
		t.Fatalf("Migrate() error = %q, want substring %q", err.Error(), "nil db")
	}
}

func sqliteObjectExists(t *testing.T, db *sql.DB, objectType, name string) bool {
	t.Helper()

	var found int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?`,
		objectType,
		name,
	).Scan(&found)
	if err != nil {
		t.Fatalf("QueryRow(sqlite_master lookup for %s %q) error = %v", objectType, name, err)
	}

	return found == 1
}

func migratedTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db := openTestDB(t)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	return db
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("PRAGMA foreign_keys = ON error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("QueryRow(%q) error = %v", query, err)
	}
	return n
}
