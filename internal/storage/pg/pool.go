package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	ConnStr string
}
type ConnectionPool struct {
	conn *pgxpool.Pool
}

func NewConnectionPool(ctx context.Context, cfg PoolConfig) (*ConnectionPool, error) {
	dbpool, err := pgxpool.New(ctx, cfg.ConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection conn: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return &ConnectionPool{conn: dbpool}, nil
}

func (p *ConnectionPool) GetConn() *pgxpool.Pool {
	return p.conn
}

func (p *ConnectionPool) Close() {
	p.conn.Close()
}

func (p *ConnectionPool) Ping(ctx context.Context) error {
	c, err := p.conn.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()
	return c.Ping(ctx)
}
