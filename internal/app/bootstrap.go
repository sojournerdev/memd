package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sojournerdev/memd/internal/db"
	"github.com/sojournerdev/memd/internal/memory"
	"github.com/sojournerdev/memd/internal/paths"
	"github.com/sojournerdev/memd/internal/store"
)

// App holds long-lived resources owned by the server process.
type App struct {
	Paths  paths.Paths
	DB     *sql.DB
	Memory *memory.Service
}

// Bootstrap resolves runtime paths and initializes the process resources they
// require.
func Bootstrap(ctx context.Context) (*App, error) {
	p, err := paths.Resolve()
	if err != nil {
		return nil, fmt.Errorf("app: resolve paths: %w", err)
	}
	return BootstrapPaths(ctx, p)
}

// BootstrapPaths initializes process resources for p and returns a ready App.
func BootstrapPaths(ctx context.Context, p paths.Paths) (*App, error) {
	dbh, err := db.Open(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("app: open db: %w", err)
	}

	if err := store.Migrate(ctx, dbh); err != nil {
		if closeErr := dbh.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close db: %w", closeErr))
		}
		return nil, fmt.Errorf("app: migrate db: %w", err)
	}

	memories, err := store.NewSQLiteRepository(dbh)
	if err != nil {
		if closeErr := dbh.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close db: %w", closeErr))
		}
		return nil, fmt.Errorf("app: create memory repository: %w", err)
	}

	memoryService, err := memory.NewService(memories)
	if err != nil {
		if closeErr := dbh.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close db: %w", closeErr))
		}
		return nil, fmt.Errorf("app: create memory service: %w", err)
	}

	return &App{
		Paths:  p,
		DB:     dbh,
		Memory: memoryService,
	}, nil
}

// Close releases the resources owned by App.
func (app *App) Close() error {
	if app == nil || app.DB == nil {
		return nil
	}
	return app.DB.Close()
}
