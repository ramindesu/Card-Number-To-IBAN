package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/config"
)

// NewPostgresPool creates and verifies a PostgreSQL connection pool.
func NewPostgresPool(ctx context.Context, databaseConfig config.Database) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}

	poolConfig.ConnConfig.ConnectTimeout = databaseConfig.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	pingContext, cancel := context.WithTimeout(ctx, databaseConfig.ConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingContext); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return pool, nil
}
