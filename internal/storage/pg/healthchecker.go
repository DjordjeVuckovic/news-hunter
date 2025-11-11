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

	err := hc.pool.Ping(ctx)
	if err != nil {
		return false
	}

	return true
}
