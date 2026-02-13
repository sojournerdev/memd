package db

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sojournerdev/memd/internal/paths"
)

func TestOpen(t *testing.T) {
	t.Run("empty DBPath returns error", func(t *testing.T) {
		t.Parallel()

		_, err := Open(context.Background(), paths.Paths{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		// Assert on a stable substring to avoid coupling to exact formatting.
		if !strings.Contains(err.Error(), "db: empty DBPath") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("creates dirs, opens db, applies pragmas", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tmp := t.TempDir()
		p := paths.Paths{
			StateDir: tmp,
			DBPath:   filepath.Join(tmp, "memd.db"),
			BlobsDir: filepath.Join(tmp, "blobs"),
		}

		db, err := Open(ctx, p)
		if err != nil {
			t.Fatalf("Open() error: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		if _, err := os.Stat(p.StateDir); err != nil {
			t.Fatalf("state dir missing: %v", err)
		}
		if _, err := os.Stat(p.BlobsDir); err != nil {
			t.Fatalf("blobs dir missing: %v", err)
		}
		// Ensures Open created/initialized a file-backed DB (not just an in-memory handle).
		if _, err := os.Stat(p.DBPath); err != nil {
			t.Fatalf("db file missing: %v", err)
		}

		if err := db.PingContext(ctx); err != nil {
			t.Fatalf("PingContext() error: %v", err)
		}

		var journalMode string
		if err := db.QueryRowContext(ctx, `PRAGMA journal_mode;`).Scan(&journalMode); err != nil {
			t.Fatalf("read journal_mode: %v", err)
		}
		if strings.ToLower(journalMode) != "wal" {
			t.Fatalf("journal_mode = %q, want wal", journalMode)
		}

		var fk int
		if err := db.QueryRowContext(ctx, `PRAGMA foreign_keys;`).Scan(&fk); err != nil {
			t.Fatalf("read foreign_keys: %v", err)
		}
		if fk != 1 {
			t.Fatalf("foreign_keys = %d, want 1", fk)
		}

		var busy int
		if err := db.QueryRowContext(ctx, `PRAGMA busy_timeout;`).Scan(&busy); err != nil {
			t.Fatalf("read busy_timeout: %v", err)
		}
		if busy != 5000 {
			t.Fatalf("busy_timeout = %d, want 5000", busy)
		}

		var sync int
		if err := db.QueryRowContext(ctx, `PRAGMA synchronous;`).Scan(&sync); err != nil {
			t.Fatalf("read synchronous: %v", err)
		}
		// Guardrail: if Open changes the durability policy, this should fail loudly.
		if sync != 1 {
			t.Fatalf("synchronous = %d, want 1 (NORMAL)", sync)
		}
	})
}

func TestWithShortTimeout(t *testing.T) {
	t.Run("uses parent deadline when sooner", func(t *testing.T) {
		t.Parallel()

		parent, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		parentDeadline, ok := parent.Deadline()
		if !ok {
			t.Fatalf("expected parent deadline")
		}

		ctx, cancel2 := withShortTimeout(parent, 5*time.Second)
		defer cancel2()

		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatalf("expected ctx deadline")
		}

		if deadline.After(parentDeadline) {
			t.Fatalf("ctx deadline %v is after parent deadline %v", deadline, parentDeadline)
		}
	})

	t.Run("adds deadline when parent has none", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := withShortTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		if _, ok := ctx.Deadline(); !ok {
			t.Fatalf("expected deadline")
		}
	})
}
