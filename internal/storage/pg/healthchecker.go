package pg

import (
	"context"
)

type HealthChecker struct {
	pool *ConnectionPool
}

func NewHealthChecker(pool *ConnectionPool) *HealthChecker {
	return &HealthChecker{
		pool: pool,
	}
}

func (hc *HealthChecker) Healthy(ctx context.Context) bool {
	if hc.pool == nil {
		return false
	}

	return hc.pool.Ping(ctx) == nil
}
