package runner

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/metrics"
	"github.com/google/uuid"
)

type QueryResult struct {
	QueryID      string
	JobName      string
	Layer        string
	EngineName   string
	Scores       metrics.ScoreSet
	RankedDocIDs []uuid.UUID
	TotalMatches int64
	Latency      LatencyStats
	Error        error
}

type JobResult struct {
	JobName     string
	Layer       string
	Results     map[string]map[string]QueryResult // [queryID][engineName]
	QueryOrder  []string
	EngineNames []string
}

type BenchmarkResult struct {
	Jobs   []*JobResult
	Config Config
}

func (br *BenchmarkResult) AllEngineNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, jr := range br.Jobs {
		for _, name := range jr.EngineNames {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}
