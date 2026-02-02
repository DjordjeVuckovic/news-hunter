package runner

import (
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/benchmark/metrics"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/query"
)

type QueryResult struct {
	QueryID      string
	QueryKind    query.Kind
	EngineName   string
	Scores       metrics.ScoreSet
	TotalMatches int64
	Latency      time.Duration
	Error        error
}

// BenchmarkResult maps queryID -> engineName -> QueryResult.
type BenchmarkResult struct {
	Results     map[string]map[string]QueryResult
	QueryOrder  []string
	EngineNames []string
	Config      Config
}
