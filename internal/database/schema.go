package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CheckSchema(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	provider, err := newGooseProvider(pool)
	if err != nil {
		return false, fmt.Errorf("migration provider: %w", err)
	}
	defer provider.Close()

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	hasPending, err := provider.HasPending(timeoutCtx)
	if err != nil {
		return false, fmt.Errorf("check pending migrations: %w", err)
	}

	return hasPending, nil
}
