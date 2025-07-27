package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	ConnStr string
}
type ConnectionPool struct {
	conn *pgxpool.Pool
}

func NewConnectionPool(ctx context.Context, cfg Config) (*ConnectionPool, error) {
	dbpool, err := pgxpool.New(ctx, cfg.ConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection conn: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return &ConnectionPool{conn: dbpool}, nil
}
