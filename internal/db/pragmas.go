package db

import (
	"context"
	"database/sql"
	"fmt"
)

const pragmaJournalModeWAL = "PRAGMA journal_mode = WAL;"

var connectionPragmas = []string{
	"PRAGMA foreign_keys = ON;",
	"PRAGMA busy_timeout = 5000;",
	"PRAGMA synchronous = NORMAL;",
}

func applyPragmas(parent context.Context, db *sql.DB) error {
	ctx, cancel := withShortTimeout(parent, dbTimeout)
	defer cancel()

	if _, err := db.ExecContext(ctx, pragmaJournalModeWAL); err != nil {
		return fmt.Errorf("db: pragma %q: %w", pragmaJournalModeWAL, err)
	}

	for _, s := range connectionPragmas {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("db: pragma %q: %w", s, err)
		}
	}

	return nil
}
