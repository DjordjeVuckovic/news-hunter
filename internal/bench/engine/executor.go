package engine

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Executor interface {
	Execute(ctx context.Context, query string, params []any) (*Execution, error)
	Name() string
	Close() error
}

type Execution struct {
	RankedDocIDs []uuid.UUID
	TotalMatches int64
	Latency      time.Duration
}
