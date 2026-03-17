package commands

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestInit_OK_Idempotent_AndCreatesSchema(t *testing.T) {
	state := t.TempDir()
	t.Setenv("MEMD_HOME", state)

	var out1, errOut1 bytes.Buffer
	got1 := Init(&out1, &errOut1)

	if got1 != ExitOK {
		t.Fatalf("Init() first run = %d, want %d; errOut=%q", got1, ExitOK, errOut1.String())
	}
	if errOut1.Len() != 0 {
		t.Fatalf("Init() first run errOut = %q, want empty", errOut1.String())
	}
	if !strings.Contains(out1.String(), "OK\n") {
		t.Fatalf("Init() first run out missing OK; got:\n%s", out1.String())
	}
	for _, want := range []string{
		"state_dir: " + state,
		"db_path:   " + filepath.Join(state, "memd.db"),
		"blobs_dir: " + filepath.Join(state, "blobs"),
		"schema:    ready",
	} {
		if !strings.Contains(out1.String(), want) {
			t.Fatalf("Init() first run out missing %q; got:\n%s", want, out1.String())
		}
	}

	dbPath := filepath.Join(state, "memd.db")
	conn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("conn.Close() error = %v", err)
		}
	})

	migrationRowsAfterFirstRun := countRows(t, conn, `SELECT COUNT(*) FROM migrations`)
	if migrationRowsAfterFirstRun == 0 {
		t.Fatal("migrations row count after first init = 0, want > 0")
	}

	var out2, errOut2 bytes.Buffer
	got2 := Init(&out2, &errOut2)
	if got2 != ExitOK {
		t.Fatalf("Init() second run = %d, want %d; errOut=%q", got2, ExitOK, errOut2.String())
	}
	if errOut2.Len() != 0 {
		t.Fatalf("Init() second run errOut = %q, want empty", errOut2.String())
	}
	migrationRowsAfterSecondRun := countRows(t, conn, `SELECT COUNT(*) FROM migrations`)
	if migrationRowsAfterSecondRun != migrationRowsAfterFirstRun {
		t.Fatalf("migrations row count after second init = %d, want %d", migrationRowsAfterSecondRun, migrationRowsAfterFirstRun)
	}

	ctx := context.Background()
	for _, table := range []string{
		"migrations",
		"repo_files",
		"repo_chunks",
		"repo_chunks_fts",
	} {
		if !tableExists(t, ctx, conn, table) {
			t.Fatalf("expected table %q to exist after init", table)
		}
	}

	var latestMigrationVersion int
	err = conn.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM migrations`).Scan(&latestMigrationVersion)
	if err != nil {
		t.Fatalf("query latest migration version error = %v", err)
	}
	if latestMigrationVersion != 1 {
		t.Fatalf("latest migration version = %d, want %d", latestMigrationVersion, 1)
	}
}

func TestInit_Error_WhenMEMD_HOMEIsFile(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(stateFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("MEMD_HOME", stateFile)

	var out, errOut bytes.Buffer
	got := Init(&out, &errOut)

	if got != ExitError {
		t.Fatalf("Init() = %d, want %d", got, ExitError)
	}
	if out.Len() != 0 {
		t.Fatalf("Init() out = %q, want empty", out.String())
	}
	if errOut.Len() == 0 {
		t.Fatal("Init() errOut is empty; want error output")
	}
	if !strings.Contains(errOut.String(), "memd: init:") {
		t.Fatalf("errOut = %q, want to contain %q", errOut.String(), "memd: init:")
	}
}

func tableExists(t *testing.T, ctx context.Context, db *sql.DB, table string) bool {
	t.Helper()

	var n int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		table,
	).Scan(&n)
	if err != nil {
		t.Fatalf("tableExists(%q) query error = %v", table, err)
	}
	return n == 1
}

func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("QueryRow(%q) error = %v", query, err)
	}
	return n
}
