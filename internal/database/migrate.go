package database

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) ([]*goose.MigrationResult, error) {
	provider, err := newGooseProvider(pool)
	if err != nil {
		return nil, fmt.Errorf("migration provider: %w", err)
	}
	defer provider.Close()

	res, err := provider.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	return res, nil
}
