package main

import (
	"context"
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
		slog.Error("memd exited", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	app, err := app.Bootstrap(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			slog.Error("shutdown failed", "err", err)
		}
	}()

	server, err := mcp.New(app.Memory, mcp.Options{
		Version: version,
	})
	if err != nil {
		return fmt.Errorf("create mcp server: %w", err)
	}

	return server.RunStdio(ctx)
}
