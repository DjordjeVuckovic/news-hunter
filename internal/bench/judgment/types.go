package judgment

import (
	"context"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"github.com/google/uuid"
)

type Judge interface {
	Grade(ctx context.Context, entry pool.PoolEntry) ([]GradedDoc, error)
}

type GradedDoc struct {
	DocID uuid.UUID `yaml:"doc_id"`
	Grade int       `yaml:"grade"`
}

type JudgmentFile struct {
	Strategy string          `yaml:"strategy"`
	Queries  []JudgmentEntry `yaml:"queries"`
}

type JudgmentEntry struct {
	QueryID string      `yaml:"query_id"`
	Docs    []GradedDoc `yaml:"docs"`
}
