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

// App holds the long-lived resources owned by the server process.
//
// It gives entrypoints a single value to pass around and close when the process
// shuts down.
type App struct {
	Paths  paths.Paths
	DB     *sql.DB
	Memory *memory.Service
}

// Bootstrap resolves the default runtime paths and builds the application.
//
// It delegates the actual resource initialization to BootstrapPaths after
// choosing the process paths used in normal startup.
func Bootstrap(ctx context.Context) (*App, error) {
	p, err := paths.Resolve()
	if err != nil {
		return nil, fmt.Errorf("app: resolve paths: %w", err)
	}
	return BootstrapPaths(ctx, p)
}

// BootstrapPaths builds the application using explicit runtime paths.
//
// It opens SQLite, runs migrations, wires the memory repository and service,
// and returns the initialized resources as an App.
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

	return &App{
		Paths:  p,
		DB:     dbh,
		Memory: memory.NewService(memories),
	}, nil
}

// Close releases the resources owned by App.
func (app *App) Close() error {
	return app.DB.Close()
}
