package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Embedded SQL migrations bundled into the binary.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version  int
	name     string
	sql      string
	checksum string
}

// Migrate applies embedded SQL migrations in order.
// It records checksums to detect edited migrations and runs everything in one transaction.
func Migrate(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("store: migrate: nil db")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	migs, err := loadMigrations()
	if err != nil {
		return err
	}
	if len(migs) == 0 {
		return fmt.Errorf("store: no embedded migrations found")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: begin migration tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := ensureMigrationsTable(ctx, tx); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, tx)
	if err != nil {
		return err
	}

	for _, m := range migs {
		if _, ok := applied[m.version]; ok {
			if err := verifyChecksum(ctx, tx, m); err != nil {
				return err
			}
			continue
		}
		if err := applyMigration(ctx, tx, m); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: commit migration run: %w", err)
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("store: read migrations dir: %w", err)
	}

	var migs []migration
	seen := map[int]string{}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		name := e.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("store: invalid migration filename %q (want NNNN_name.sql)", name)
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("store: invalid migration version in %q: %w", name, err)
		}
		if prev, ok := seen[version]; ok {
			return nil, fmt.Errorf("store: duplicate migration version %04d: %q and %q", version, prev, name)
		}
		seen[version] = name

		b, err := migrationsFS.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			return nil, fmt.Errorf("store: read migration %s: %w", name, err)
		}

		sqlText := string(b)
		sum := sha256.Sum256([]byte(sqlText))

		migs = append(migs, migration{
			version:  version,
			name:     name,
			sql:      sqlText,
			checksum: hex.EncodeToString(sum[:]),
		})
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })
	return migs, nil
}

func ensureMigrationsTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY CHECK (version > 0),
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			applied_at_ns INTEGER NOT NULL CHECK (applied_at_ns > 0),
			CHECK (length(trim(name)) > 0),
			CHECK (length(trim(checksum)) > 0)
		) STRICT, WITHOUT ROWID;
	`)
	if err != nil {
		return fmt.Errorf("store: ensure migrations table: %w", err)
	}
	return nil
}

func appliedVersions(ctx context.Context, tx *sql.Tx) (map[int]struct{}, error) {
	rows, err := tx.QueryContext(ctx, `SELECT version FROM migrations`)
	if err != nil {
		return nil, fmt.Errorf("store: query migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[int]struct{}{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("store: scan migration version: %w", err)
		}
		out[v] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate migrations: %w", err)
	}
	return out, nil
}

func verifyChecksum(ctx context.Context, tx *sql.Tx, m migration) error {
	var recorded string
	err := tx.QueryRowContext(ctx, `SELECT checksum FROM migrations WHERE version = ?`, m.version).Scan(&recorded)
	if err != nil {
		return fmt.Errorf("store: read checksum for migration %d (%s): %w", m.version, m.name, err)
	}
	if recorded != m.checksum {
		return fmt.Errorf(
			"store: migration checksum mismatch for version %d (%s): db=%s file=%s (did you edit an applied migration?)",
			m.version, m.name, recorded, m.checksum,
		)
	}
	return nil
}

func applyMigration(ctx context.Context, tx *sql.Tx, m migration) error {
	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return fmt.Errorf("store: exec migration %d (%s): %w", m.version, m.name, err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO migrations(version, name, checksum, applied_at_ns) VALUES(?, ?, ?, ?)`,
		m.version, m.name, m.checksum, time.Now().UnixNano(),
	); err != nil {
		return fmt.Errorf("store: record migration %d (%s): %w", m.version, m.name, err)
	}

	return nil
}
