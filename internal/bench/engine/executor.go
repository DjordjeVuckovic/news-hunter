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

// Validator is an optional capability for executors that can syntactically
// validate a query without executing it. PG uses EXPLAIN, ES uses
// _validate/query, API parses the request descriptor. The CLI's `validate`
// subcommand uses this to fail fast on broken queries before a real run.
type Validator interface {
	Validate(ctx context.Context, query string) error
}

type Execution struct {
	RankedDocIDs []uuid.UUID
	TotalMatches int64
	Latency      time.Duration
}
