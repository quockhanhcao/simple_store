package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	cfg, err := config.Load(getenv, ":5200")
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

	// check pending migrations
	pendingMigrations, err := database.CheckSchema(ctx, pool)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if pendingMigrations {
		return errors.New("database schema is out of date: run 'make migrate'")
	}

	// handler
	srv := NewHandler(
		logger,
		cfg,
	)
	//server
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// buffer size 1
	// goroutine never blocks on send
	serverErr := make(chan error, 1)
	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	logger.Info("server listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
	go func() {
		serverErr <- httpServer.Serve(listener)
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		logger.Info("shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		logger.Info("server stopped")
		return nil
	}
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
