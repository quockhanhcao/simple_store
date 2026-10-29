package database

import (
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func newGooseProvider(pool *pgxpool.Pool) (*goose.Provider, error) {
	migrations, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create migration provider: %w", err)
	}
	return provider, nil
}
