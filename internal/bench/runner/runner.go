package runner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/metrics"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
)

type Runner struct {
	suite   *suite.TestSuite
	engines []engine.SearchEngine
	config  Config
}

func New(s *suite.TestSuite, engines []engine.SearchEngine, cfg Config) *Runner {
	return &Runner{
		suite:   s,
		engines: engines,
		config:  cfg,
	}
}

func (r *Runner) Run(ctx context.Context) (*BenchmarkResult, error) {
	if len(r.engines) == 0 {
		return nil, fmt.Errorf("no engines configured")
	}

	engineNames := make([]string, len(r.engines))
	for i, eng := range r.engines {
		engineNames[i] = eng.Name
	}

	br := &BenchmarkResult{
		Results:     make(map[string]map[string]QueryResult),
		QueryOrder:  make([]string, 0, len(r.suite.Queries)),
		EngineNames: engineNames,
		Config:      r.config,
	}

	for i := range r.suite.Queries {
		bq := &r.suite.Queries[i]
		br.QueryOrder = append(br.QueryOrder, bq.ID)
		br.Results[bq.ID] = make(map[string]QueryResult, len(r.engines))

		judgments := bq.JudgmentMap()

		for _, eng := range r.engines {
			exec := engine.Execute(ctx, eng, bq, r.config.MaxK)

			var scores metrics.ScoreSet
			if exec.Error == nil && len(judgments) > 0 {
				scores = metrics.ComputeAll(exec.RankedDocIDs, judgments, r.config.KValues, r.config.RelevanceThreshold)
			}

			qr := QueryResult{
				QueryID:      bq.ID,
				QueryKind:    bq.Kind,
				EngineName:   eng.Name,
				Scores:       scores,
				TotalMatches: exec.TotalMatches,
				Latency:      exec.Latency,
				Error:        exec.Error,
			}

			br.Results[bq.ID][eng.Name] = qr

			if exec.Error != nil {
				slog.Warn("query execution failed", "query", bq.ID, "engine", eng.Name, "error", exec.Error)
			}
		}
	}

	return br, nil
}
