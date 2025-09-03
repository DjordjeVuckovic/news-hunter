package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthChecker struct {
	pg *pgxpool.Conn
}

func NewHealthChecker(pg *pgxpool.Conn) *HealthChecker {
	return &HealthChecker{
		pg: pg,
	}
}

func (hc *HealthChecker) Healthy(ctx context.Context) bool {
	if hc.pg == nil {
		return false
	}

	err := hc.pg.Ping(ctx)
	if err != nil {
		return false
	}

	return true
}
