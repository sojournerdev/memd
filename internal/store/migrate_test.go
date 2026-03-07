package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestParseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filename    string
		want        int
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid",
			filename: "0001_init.sql",
			want:     1,
		},
		{
			name:        "missing underscore",
			filename:    "0001init.sql",
			wantErr:     true,
			errContains: "invalid migration filename",
		},
		{
			name:        "non numeric version",
			filename:    "abcd_init.sql",
			wantErr:     true,
			errContains: "invalid migration version",
		},
		{
			name:        "empty version",
			filename:    "_init.sql",
			wantErr:     true,
			errContains: "invalid migration version",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseVersion(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseVersion(%q) error = nil, want error", tt.filename)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("parseVersion(%q) error = %q, want substring %q", tt.filename, err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseVersion(%q) error = %v, want nil", tt.filename, err)
			}
			if got != tt.want {
				t.Fatalf("parseVersion(%q) = %d, want %d", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMigrate_AppliesAllEmbeddedMigrations(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("loadMigrations() returned no migrations")
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	got := countRows(t, db, `SELECT COUNT(*) FROM migrations`)
	if got != len(migs) {
		t.Fatalf("migration row count = %d, want %d", got, len(migs))
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("loadMigrations() returned no migrations")
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() first run error = %v", err)
	}
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() second run error = %v", err)
	}

	got := countRows(t, db, `SELECT COUNT(*) FROM migrations`)
	if got != len(migs) {
		t.Fatalf("migration row count after second run = %d, want %d", got, len(migs))
	}
}

func TestMigrate_RecordsChecksums(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("loadMigrations() returned no migrations")
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	got := countRows(t, db, `SELECT COUNT(*) FROM migrations WHERE checksum <> ''`)
	if got != len(migs) {
		t.Fatalf("non-empty checksum row count = %d, want %d", got, len(migs))
	}
}

func TestMigrate_FailsOnChecksumMismatch(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("loadMigrations() returned no migrations")
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() initial run error = %v", err)
	}

	_, err = db.Exec(
		`UPDATE migrations SET checksum = ? WHERE version = ?`,
		"not-the-real-checksum",
		migs[0].version,
	)
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

func TestMigrate_NilContext(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	if err := Migrate(nil, db); err != nil {
		t.Fatalf("Migrate(nil, db) error = %v", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
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
