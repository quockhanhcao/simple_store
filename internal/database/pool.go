package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnectionPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(timeoutCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
