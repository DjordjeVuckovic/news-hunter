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

	for _, job := range bs.Jobs {
		loaded, err := suite.LoadFromFile(job.Suite)
		if err != nil {
			return nil, fmt.Errorf("load suite for job %q: %w", job.Name, err)
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
		judgments := q.JudgmentMap()

		for engName, exec := range executors {
			resolved, err := q.ResolveEngineQuery(engName, registry, suiteDir)
			if err != nil {
				qr := QueryResult{
					QueryID:    q.ID,
					JobName:    jr.JobName,
					EngineName: engName,
					Error:      fmt.Errorf("resolve query: %w", err),
				}
				jr.Results[q.ID][engName] = qr
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

			qr := QueryResult{
				QueryID:      q.ID,
				JobName:      jr.JobName,
				EngineName:   engName,
				Scores:       scores,
				RankedDocIDs: result.rankedIDs,
				TotalMatches: result.totalMatches,
				Latency:      result.latencyStats,
				Error:        result.err,
			}
			jr.Results[q.ID][engName] = qr

			if result.err != nil {
				slog.Warn("query failed", "query", q.ID, "engine", engName, "error", result.err)
			}
		}
	}
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
