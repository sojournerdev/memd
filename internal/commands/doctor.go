package commands

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sojournerdev/memd/internal/db"
	"github.com/sojournerdev/memd/internal/paths"
)

const doctorTimeout = 3 * time.Second

func Doctor(out, errOut io.Writer) int {
	p, err := paths.Resolve()
	if err != nil {
		tryWritef(errOut, "memd: doctor: resolve paths: %v\n", err)
		return ExitError
	}

	ctx, cancel := context.WithTimeout(context.Background(), doctorTimeout)
	defer cancel()

	conn, err := db.Open(ctx, p)
	if err != nil {
		tryWritef(errOut, "memd: doctor: %v\n", err)
		return ExitError
	}
	// Close is best-effort for this short-lived CLI command.
	defer func() { _ = conn.Close() }()

	var journalMode string
	if err := conn.QueryRowContext(ctx, `PRAGMA journal_mode;`).Scan(&journalMode); err != nil {
		tryWritef(errOut, "memd: doctor: read journal_mode: %v\n", err)
		return ExitError
	}

	var sqliteVer string
	if err := conn.QueryRowContext(ctx, `SELECT sqlite_version();`).Scan(&sqliteVer); err != nil {
		tryWritef(errOut, "memd: doctor: read sqlite_version: %v\n", err)
		return ExitError
	}

	var fk int
	if err := conn.QueryRowContext(ctx, `PRAGMA foreign_keys;`).Scan(&fk); err != nil {
		tryWritef(errOut, "memd: doctor: read foreign_keys: %v\n", err)
		return ExitError
	}

	var busy int
	if err := conn.QueryRowContext(ctx, `PRAGMA busy_timeout;`).Scan(&busy); err != nil {
		tryWritef(errOut, "memd: doctor: read busy_timeout: %v\n", err)
		return ExitError
	}

	var sync int
	if err := conn.QueryRowContext(ctx, `PRAGMA synchronous;`).Scan(&sync); err != nil {
		tryWritef(errOut, "memd: doctor: read synchronous: %v\n", err)
		return ExitError
	}

	dbWritable := false
	// Probe write access in a transaction and roll it back.
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		tryWritef(errOut, "memd: doctor: begin tx: %v\n", err)
		return ExitError
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS __memd_doctor_tmp (id INTEGER);`); err != nil {
		_ = tx.Rollback()
		tryWritef(errOut, "memd: doctor: write test: %v\n", err)
		return ExitError
	}
	_ = tx.Rollback()
	dbWritable = true

	if err := writef(
		out,
		"OK\nstate_dir:    %s\ndb_path:      %s\nblobs_dir:    %s\njournal_mode: %s\nsqlite_ver:   %s\nforeign_keys: %t\nbusy_timeout: %dms\nsynchronous:  %s (%d)\ndb_writable:  %t\n",
		p.StateDir,
		p.DBPath,
		p.BlobsDir,
		journalMode,
		sqliteVer,
		fk == 1,
		busy,
		synchronousLabel(sync),
		sync,
		dbWritable,
	); err != nil {
		return ExitError
	}

	return ExitOK
}

func synchronousLabel(v int) string {
	switch v {
	case 0:
		return "OFF"
	case 1:
		return "NORMAL"
	case 2:
		return "FULL"
	case 3:
		return "EXTRA"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", v)
	}
}
