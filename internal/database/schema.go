package database

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func CheckSchema(ctx context.Context, pool *pgxpool.Pool) (bool, error) {

	migrations, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return false, fmt.Errorf("open embedded migrations: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		return false, fmt.Errorf("create migration provider: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	hasPending, err := provider.HasPending(timeoutCtx)
	if err != nil {
		return false, fmt.Errorf("check pending migrations: %w", err)
	}

	return hasPending, nil
}
