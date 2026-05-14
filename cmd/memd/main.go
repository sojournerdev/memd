package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sojournerdev/memd/internal/app"
	"github.com/sojournerdev/memd/internal/mcp"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Info("memd stopped", "err", err)
			return
		}
		slog.Error("memd exited", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	slog.Info("memd starting", "version", version)

	app, err := app.Bootstrap(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			slog.Error("shutdown failed", "err", err)
		}
	}()

	srv := mcp.New(app.Memory, mcp.Options{
		Version: version,
	})

	if err := srv.RunStdio(ctx); err != nil {
		return fmt.Errorf("mcp stdio: %w", err)
	}

	return nil
}
