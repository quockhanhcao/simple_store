package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("open embedded migrations: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	_, err = provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
