package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/metrics"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
	"github.com/google/uuid"
)

type Runner struct {
	config Config
}

func New(cfg Config) *Runner {
	return &Runner{config: cfg}
}

func (r *Runner) RunAll(
	ctx context.Context,
	bs *spec.BenchSpec,
	executors map[string]engine.Executor,
) (*BenchmarkResult, error) {
	br := &BenchmarkResult{Config: r.config}

	// Cache suite loads — multiple jobs commonly share a suite.
	suiteCache := map[string]*suite.LoadedSuite{}
	for _, job := range bs.Jobs {
		loaded, ok := suiteCache[job.Suite]
		if !ok {
			ls, err := suite.LoadFromFile(job.Suite)
			if err != nil {
				return nil, fmt.Errorf("load suite for job %q: %w", job.Name, err)
			}
			suiteCache[job.Suite] = ls
			loaded = ls
		}

		jr, err := r.RunJob(ctx, job, loaded, executors)
		if err != nil {
			return nil, fmt.Errorf("run job %q: %w", job.Name, err)
		}
		br.Jobs = append(br.Jobs, jr)
	}

	return br, nil
}

func (r *Runner) RunJob(
	ctx context.Context,
	job spec.Job,
	loaded *suite.LoadedSuite,
	executors map[string]engine.Executor,
) (*JobResult, error) {
	jobExecutors := make(map[string]engine.Executor)
	for _, engName := range job.Engines {
		exec, ok := executors[engName]
		if !ok {
			return nil, fmt.Errorf("executor %q not found", engName)
		}
		jobExecutors[engName] = exec
	}

	jr := &JobResult{
		JobName:     job.Name,
		Results:     make(map[string]map[string]QueryResult),
		EngineNames: job.Engines,
	}

	r.runQueries(ctx, jr, loaded.Suite.Queries, loaded.Registry, jobExecutors, loaded.Dir)

	return jr, nil
}

func (r *Runner) runQueries(
	ctx context.Context,
	jr *JobResult,
	queries []suite.Query,
	registry *suite.TemplateRegistry,
	executors map[string]engine.Executor,
	suiteDir string,
) {
	for i := range queries {
		q := &queries[i]
		jr.QueryOrder = append(jr.QueryOrder, q.ID)
		jr.Results[q.ID] = make(map[string]QueryResult)
		judgments := r.judgmentsFor(q)

		for _, engName := range jr.EngineNames {
			exec, ok := executors[engName]
			if !ok {
				continue
			}
			resolved, err := q.ResolveEngineQuery(engName, registry, suiteDir)
			if err != nil {
				jr.Results[q.ID][engName] = QueryResult{
					QueryID:    q.ID,
					EngineName: engName,
					Error:      fmt.Errorf("resolve query: %w", err),
				}
				slog.Warn("resolve query failed", "query", q.ID, "engine", engName, "error", err)
				continue
			}
			if resolved == nil {
				continue
			}

			result := r.executeWithRetries(ctx, exec, resolved.Query, nil, r.config.WarmupRuns, r.config.Runs)

			var scores metrics.ScoreSet
			if result.err == nil && len(judgments) > 0 {
				scores = metrics.ComputeAll(result.rankedIDs, judgments, r.config.KValues, r.config.RelevanceThreshold)
			}

			jr.Results[q.ID][engName] = QueryResult{
				QueryID:      q.ID,
				EngineName:   engName,
				Scores:       scores,
				RankedDocIDs: result.rankedIDs,
				TotalMatches: result.totalMatches,
				Latency:      result.latencyStats,
				Error:        result.err,
			}

			if result.err != nil {
				slog.Warn("query failed", "query", q.ID, "engine", engName, "error", result.err)
			}
		}
	}
}

// judgmentsFor returns the relevance grades for a query. Priority: the
// runner-level Config.Judgments map (loaded by the CLI from the resolved
// annotations file) takes precedence over any judgments embedded in the suite
// (which is the case only when a suite is hand-edited — rare in v1).
func (r *Runner) judgmentsFor(q *suite.Query) map[uuid.UUID]int {
	if r.config.Judgments != nil {
		if perQuery, ok := r.config.Judgments[q.ID]; ok {
			out := make(map[uuid.UUID]int, len(perQuery))
			for idStr, grade := range perQuery {
				if id, err := uuid.Parse(idStr); err == nil {
					out[id] = grade
				}
			}
			return out
		}
	}
	return q.JudgmentMap()
}

type execResult struct {
	rankedIDs    []uuid.UUID
	totalMatches int64
	latencyStats LatencyStats
	err          error
}

func (r *Runner) executeWithRetries(
	ctx context.Context,
	exec engine.Executor,
	query string,
	params []any,
	warmup, runs int,
) execResult {
	for i := 0; i < warmup; i++ {
		_, _ = exec.Execute(ctx, query, params)
	}

	var latencies []time.Duration
	var lastExec *engine.Execution
	var lastErr error

	for i := 0; i < runs; i++ {
		result, err := exec.Execute(ctx, query, params)
		if err != nil {
			lastErr = err
			continue
		}
		lastExec = result
		latencies = append(latencies, result.Latency)
	}

	if lastExec == nil {
		return execResult{err: lastErr}
	}

	return execResult{
		rankedIDs:    lastExec.RankedDocIDs,
		totalMatches: lastExec.TotalMatches,
		latencyStats: ComputeLatencyStats(latencies),
	}
}
