package commands

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sojournerdev/memd/internal/db"
	"github.com/sojournerdev/memd/internal/paths"
)

func Doctor(out, errOut io.Writer) int {
	p, err := paths.Resolve()
	if err != nil {
		fmt.Fprintf(errOut, "memd: doctor: resolve paths: %v\n", err)
		return ExitError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := db.Open(ctx, p)
	if err != nil {
		fmt.Fprintf(errOut, "memd: doctor: %v\n", err)
		return ExitError
	}
	defer conn.Close()

	var journalMode string
	if err := conn.QueryRowContext(ctx, `PRAGMA journal_mode;`).Scan(&journalMode); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: read journal_mode: %v\n", err)
		return ExitError
	}

	var sqliteVer string
	if err := conn.QueryRowContext(ctx, `SELECT sqlite_version();`).Scan(&sqliteVer); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: read sqlite_version: %v\n", err)
		return ExitError
	}

	var fk int
	if err := conn.QueryRowContext(ctx, `PRAGMA foreign_keys;`).Scan(&fk); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: read foreign_keys: %v\n", err)
		return ExitError
	}

	var busy int
	if err := conn.QueryRowContext(ctx, `PRAGMA busy_timeout;`).Scan(&busy); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: read busy_timeout: %v\n", err)
		return ExitError
	}

	var sync int
	if err := conn.QueryRowContext(ctx, `PRAGMA synchronous;`).Scan(&sync); err != nil {
		fmt.Fprintf(errOut, "memd: doctor: read synchronous: %v\n", err)
		return ExitError
	}

	dbWritable := false
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintf(errOut, "memd: doctor: begin tx: %v\n", err)
		return ExitError
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS __memd_doctor_tmp (id INTEGER);`); err != nil {
		_ = tx.Rollback()
		fmt.Fprintf(errOut, "memd: doctor: write test: %v\n", err)
		return ExitError
	}
	_ = tx.Rollback()
	dbWritable = true

	fmt.Fprintln(out, "OK")
	fmt.Fprintf(out, "state_dir:    %s\n", p.StateDir)
	fmt.Fprintf(out, "db_path:      %s\n", p.DBPath)
	fmt.Fprintf(out, "blobs_dir:    %s\n", p.BlobsDir)
	fmt.Fprintf(out, "journal_mode: %s\n", journalMode)
	fmt.Fprintf(out, "sqlite_ver:   %s\n", sqliteVer)
	fmt.Fprintf(out, "foreign_keys: %t\n", fk == 1)
	fmt.Fprintf(out, "busy_timeout: %dms\n", busy)
	fmt.Fprintf(out, "synchronous:  %s (%d)\n", synchronousLabel(sync), sync)
	fmt.Fprintf(out, "db_writable:  %t\n", dbWritable)

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
