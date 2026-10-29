package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/quockhanhcao/simple_store/internal/config"
	"github.com/quockhanhcao/simple_store/internal/database"
	"github.com/quockhanhcao/simple_store/internal/platform"
)

func run(
	ctx context.Context,
	getenv func(string) string,
	stdout io.Writer,
) error {
	// load config
	// no server is created => no default address
	cfg, err := config.Load(getenv, "")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// create logger
	logger := platform.NewLogger(stdout, cfg.LogLevel, cfg.LogFormat)

	// db pool
	pool, err := database.NewConnectionPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	// migrate
	migrateRes, err := database.Migrate(ctx, pool)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if len(migrateRes) == 0 {
		logger.Info("schema up to date")
	} else {
		logger.Info("migrations applied", "count", len(migrateRes), "results", migrateRes)
	}

	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
