package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

type PoolConfig struct {
	ConnStr          string
	RegisterVecTypes bool
}
type ConnectionPool struct {
	conn *pgxpool.Pool
}

func NewConnectionPool(ctx context.Context, cfg PoolConfig) (*ConnectionPool, error) {
	config, err := pgxpool.ParseConfig(cfg.ConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.AfterConnect = afterConnect(cfg.RegisterVecTypes)

	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return &ConnectionPool{conn: dbpool}, nil
}

func afterConnect(registerVec bool) func(ctx context.Context, conn *pgx.Conn) error {
	return func(ctx context.Context, conn *pgx.Conn) error {
		if registerVec {
			err := pgxvec.RegisterTypes(ctx, conn)
			if err != nil {
				return err
			}
		}
		return nil
	}
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
