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
	apiExec *engine.APIExecutor,
) (*BenchmarkResult, error) {
	br := &BenchmarkResult{Config: r.config}

	for _, job := range bs.Jobs {
		loaded, err := suite.LoadFromFile(job.Suite)
		if err != nil {
			return nil, fmt.Errorf("load suite for job %q: %w", job.Name, err)
		}

		jr, err := r.RunJob(ctx, job, loaded, executors, apiExec)
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
	apiExec *engine.APIExecutor,
) (*JobResult, error) {
	jobExecutors := make(map[string]engine.Executor)
	for _, engName := range job.Engines {
		exec, ok := executors[engName]
		if !ok {
			return nil, fmt.Errorf("executor %q not found", engName)
		}
		jobExecutors[engName] = exec
	}

	layer := Layer(job.Layer)
	jr := &JobResult{
		JobName:     job.Name,
		Layer:       job.Layer,
		Results:     make(map[string]map[string]QueryResult),
		EngineNames: job.Engines,
	}

	if layer == LayerRaw || layer == LayerAll {
		r.runRawQueries(ctx, jr, loaded.Suite.RawQueries, loaded.Registry, jobExecutors)
	}

	if layer == LayerAPI || layer == LayerAll {
		if apiExec == nil {
			slog.Warn("api layer requested but no API executor configured", "job", job.Name)
		} else {
			r.runAPIQueries(ctx, jr, loaded.Suite.APIQueries, apiExec)
		}
	}

	return jr, nil
}

func (r *Runner) runRawQueries(
	ctx context.Context,
	jr *JobResult,
	queries []suite.RawQuery,
	registry *suite.TemplateRegistry,
	executors map[string]engine.Executor,
) {
	for i := range queries {
		rq := &queries[i]
		jr.QueryOrder = append(jr.QueryOrder, rq.ID)
		jr.Results[rq.ID] = make(map[string]QueryResult)
		judgments := rq.JudgmentMap()

		for engName, exec := range executors {
			rawSQL, err := rq.ResolveEngineQuery(engName, registry)
			if err != nil {
				qr := QueryResult{
					QueryID:    rq.ID,
					JobName:    jr.JobName,
					Layer:      "raw",
					EngineName: engName,
					Error:      fmt.Errorf("resolve query: %w", err),
				}
				jr.Results[rq.ID][engName] = qr
				slog.Warn("resolve query failed", "query", rq.ID, "engine", engName, "error", err)
				continue
			}
			if rawSQL == "" {
				continue
			}

			result := r.executeWithRetries(ctx, exec, rawSQL, r.config.WarmupRuns, r.config.Runs)

			var scores metrics.ScoreSet
			if result.err == nil && len(judgments) > 0 {
				scores = metrics.ComputeAll(result.rankedIDs, judgments, r.config.KValues, r.config.RelevanceThreshold)
			}

			qr := QueryResult{
				QueryID:      rq.ID,
				JobName:      jr.JobName,
				Layer:        "raw",
				EngineName:   engName,
				Scores:       scores,
				RankedDocIDs: result.rankedIDs,
				TotalMatches: result.totalMatches,
				Latency:      result.latencyStats,
				Error:        result.err,
			}
			jr.Results[rq.ID][engName] = qr

			if result.err != nil {
				slog.Warn("raw query failed", "query", rq.ID, "engine", engName, "error", result.err)
			}
		}
	}
}

func (r *Runner) runAPIQueries(
	ctx context.Context,
	jr *JobResult,
	queries []suite.APIQuery,
	apiExec *engine.APIExecutor,
) {
	for i := range queries {
		aq := &queries[i]
		jr.QueryOrder = append(jr.QueryOrder, aq.ID)
		jr.Results[aq.ID] = make(map[string]QueryResult)
		judgments := aq.JudgmentMap()

		for _, backend := range aq.Backends {
			result := r.executeAPIWithRetries(ctx, apiExec, aq, r.config.WarmupRuns, r.config.Runs)

			var scores metrics.ScoreSet
			if result.err == nil && len(judgments) > 0 {
				scores = metrics.ComputeAll(result.rankedIDs, judgments, r.config.KValues, r.config.RelevanceThreshold)
			}

			qr := QueryResult{
				QueryID:      aq.ID,
				JobName:      jr.JobName,
				Layer:        "api",
				EngineName:   backend,
				Scores:       scores,
				RankedDocIDs: result.rankedIDs,
				TotalMatches: result.totalMatches,
				Latency:      result.latencyStats,
				Error:        result.err,
			}
			jr.Results[aq.ID][backend] = qr

			if result.err != nil {
				slog.Warn("api query failed", "query", aq.ID, "backend", backend, "error", result.err)
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
	rawQuery string,
	warmup, runs int,
) execResult {
	for i := 0; i < warmup; i++ {
		_, _ = exec.Execute(ctx, rawQuery)
	}

	var latencies []time.Duration
	var lastExec *engine.Execution
	var lastErr error

	for i := 0; i < runs; i++ {
		result, err := exec.Execute(ctx, rawQuery)
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

func (r *Runner) executeAPIWithRetries(
	ctx context.Context,
	apiExec *engine.APIExecutor,
	aq *suite.APIQuery,
	warmup, runs int,
) execResult {
	for i := 0; i < warmup; i++ {
		_, _ = apiExec.ExecuteAPI(ctx, aq)
	}

	var latencies []time.Duration
	var lastExec *engine.Execution
	var lastErr error

	for i := 0; i < runs; i++ {
		result, err := apiExec.ExecuteAPI(ctx, aq)
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
