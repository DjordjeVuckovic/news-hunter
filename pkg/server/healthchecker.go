package server

import "context"

type HealthChecker interface {
	Healthy(ctx context.Context) bool
}

type OkHealthChecker struct {
}

func NewOkHealthChecker() *OkHealthChecker {
	return &OkHealthChecker{}
}

func (hc *OkHealthChecker) Healthy(ctx context.Context) bool {
	return true
}
