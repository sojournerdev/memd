package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sojournerdev/memd/internal/paths"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, p paths.Paths) (*sql.DB, error) {
	if p.DBPath == "" {
		return nil, errors.New("db: empty DBPath")
	}

	if err := p.Ensure(); err != nil {
		return nil, fmt.Errorf("db: ensure paths: %w", err)
	}
	if err := p.ValidateReadWrite(); err != nil {
		return nil, fmt.Errorf("db: validate read/write: %w", err)
	}

	dsn := "file:" + p.DBPath

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := ping(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applyPragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ping(parent context.Context, db *sql.DB) error {
	ctx, cancel := withShortTimeout(parent, 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("db: ping: %w", err)
	}
	return nil
}

func applyPragmas(parent context.Context, db *sql.DB) error {
	ctx, cancel := withShortTimeout(parent, 3*time.Second)
	defer cancel()

	stmts := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
	}

	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("db: pragma %q: %w", s, err)
		}
	}
	return nil
}

func withShortTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if deadline, ok := parent.Deadline(); ok && time.Until(deadline) < d {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, d)
}
