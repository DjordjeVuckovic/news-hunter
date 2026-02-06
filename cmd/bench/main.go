package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/report"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
)

func main() {
	cfg := parseFlags()
	ctx := context.Background()

	switch cfg.Mode {
	case "bench":
		runBench(ctx, cfg)
	case "pool":
		runPool(ctx, cfg)
	case "judge":
		runJudge(cfg)
	default:
		slog.Error("Unknown mode", "mode", cfg.Mode)
		os.Exit(1)
	}
}

func runBench(ctx context.Context, cfg cliConfig) {
	kValues, err := cfg.parseKValues()
	if err != nil {
		slog.Error("Invalid k values", "error", err)
		os.Exit(1)
	}

	runCfg := runner.Config{
		KValues:            kValues,
		MaxK:               cfg.MaxK,
		RelevanceThreshold: runner.DefaultRelevanceThreshold,
		WarmupRuns:         cfg.Warmup,
		Runs:               max(cfg.Runs, 1),
	}

	if cfg.SpecPath != "" {
		runWithSpec(ctx, cfg, runCfg)
	} else {
		runQuickMode(ctx, cfg, runCfg)
	}
}

func runWithSpec(ctx context.Context, cfg cliConfig, runCfg runner.Config) {
	bs, err := spec.LoadFromFile(cfg.SpecPath)
	if err != nil {
		slog.Error("Failed to load spec", "path", cfg.SpecPath, "error", err)
		os.Exit(1)
	}

	if bs.Runs.Warmup > 0 && cfg.Warmup == 0 {
		runCfg.WarmupRuns = bs.Runs.Warmup
	}
	if bs.Runs.Iterations > 0 && cfg.Runs <= 1 {
		runCfg.Runs = bs.Runs.Iterations
	}
	if len(bs.Metrics.KValues) > 0 {
		runCfg.KValues = bs.Metrics.KValues
	}
	if bs.Metrics.MaxK > 0 {
		runCfg.MaxK = bs.Metrics.MaxK
	}
	if bs.Metrics.RelevanceThreshold > 0 {
		runCfg.RelevanceThreshold = bs.Metrics.RelevanceThreshold
	}

	executors, cleanup, err := engine.CreateFromSpec(ctx, bs.Engines)
	if err != nil {
		slog.Error("Failed to create executors", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	var apiExec *engine.APIExecutor
	if bs.API != nil && bs.API.BaseURL != "" {
		apiExec = engine.NewAPIExecutor(bs.API.BaseURL)
	}

	r := runner.New(runCfg)
	result, err := r.RunAll(ctx, bs, executors, apiExec)
	if err != nil {
		slog.Error("Benchmark failed", "error", err)
		os.Exit(1)
	}

	outputReport(result, cfg.Output)
}

func runQuickMode(ctx context.Context, cfg cliConfig, runCfg runner.Config) {
	if cfg.PgConnStr == "" && cfg.EsAddresses == "" {
		slog.Error("Quick mode requires --pg and/or --es-addresses")
		os.Exit(1)
	}

	engines := make(map[string]spec.Engine)
	var engineNames []string

	if cfg.PgConnStr != "" {
		engines["pg-native"] = spec.Engine{Type: "postgres", Connection: cfg.PgConnStr}
		engineNames = append(engineNames, "pg-native")
	}
	if cfg.EsAddresses != "" {
		engines["elasticsearch"] = spec.Engine{Type: "elasticsearch", Connection: cfg.EsAddresses, Index: cfg.EsIndex}
		engineNames = append(engineNames, "elasticsearch")
	}

	executors, cleanup, err := engine.CreateFromSpec(ctx, engines)
	if err != nil {
		slog.Error("Failed to create executors", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	var apiExec *engine.APIExecutor
	if cfg.APIURL != "" {
		apiExec = engine.NewAPIExecutor(cfg.APIURL)
	}

	bs := &spec.BenchSpec{
		Jobs: []spec.Job{
			{
				Name:    "quick",
				Suite:   cfg.SuitePath,
				Engines: engineNames,
				Layer:   cfg.Layer,
			},
		},
		Engines: engines,
	}

	r := runner.New(runCfg)
	result, err := r.RunAll(ctx, bs, executors, apiExec)
	if err != nil {
		slog.Error("Benchmark failed", "error", err)
		os.Exit(1)
	}

	outputReport(result, cfg.Output)
}

func outputReport(result *runner.BenchmarkResult, outputPath string) {
	rpt := report.Generate(result)
	report.WriteTable(rpt, os.Stdout)

	if outputPath != "" {
		if err := report.WriteJSON(rpt, outputPath); err != nil {
			slog.Error("Failed to write JSON report", "error", err)
			os.Exit(1)
		}
		slog.Info("Report written", "path", outputPath)
	}
}

func runPool(ctx context.Context, cfg cliConfig) {
	if cfg.Output == "" {
		slog.Error("Pool mode requires --output")
		os.Exit(1)
	}

	kValues, err := cfg.parseKValues()
	if err != nil {
		slog.Error("Invalid k values", "error", err)
		os.Exit(1)
	}

	runCfg := runner.Config{
		KValues:            kValues,
		MaxK:               cfg.MaxK,
		RelevanceThreshold: runner.DefaultRelevanceThreshold,
		Runs:               1,
	}

	var bs *spec.BenchSpec
	if cfg.SpecPath != "" {
		bs, err = spec.LoadFromFile(cfg.SpecPath)
		if err != nil {
			slog.Error("Failed to load spec", "error", err)
			os.Exit(1)
		}
	} else {
		bs = buildQuickSpec(cfg)
	}

	executors, cleanup, err := engine.CreateFromSpec(ctx, bs.Engines)
	if err != nil {
		slog.Error("Failed to create executors", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	r := runner.New(runCfg)
	result, err := r.RunAll(ctx, bs, executors, nil)
	if err != nil {
		slog.Error("Pool run failed", "error", err)
		os.Exit(1)
	}

	pf := buildPoolFile(result)
	if err := pool.WritePoolFile(pf, cfg.Output); err != nil {
		slog.Error("Failed to write pool file", "error", err)
		os.Exit(1)
	}
	slog.Info("Pool file written", "path", cfg.Output)
}

func buildPoolFile(result *runner.BenchmarkResult) *pool.PoolFile {
	pf := &pool.PoolFile{}
	for _, jr := range result.Jobs {
		pf.SuiteName = jr.JobName
		for _, qID := range jr.QueryOrder {
			engResults := jr.Results[qID]
			executions := make(map[string]*engine.Execution)
			for engName, qr := range engResults {
				if qr.Error != nil {
					continue
				}
				executions[engName] = &engine.Execution{
					RankedDocIDs: qr.RankedDocIDs,
					TotalMatches: qr.TotalMatches,
				}
			}
			docs := pool.PoolResults(executions, result.Config.MaxK)
			pf.Queries = append(pf.Queries, pool.PoolEntry{
				QueryID: qID,
				Docs:    docs,
			})
		}
	}
	return pf
}

func runJudge(cfg cliConfig) {
	if cfg.PoolPath == "" {
		slog.Error("Judge mode requires --pool")
		os.Exit(1)
	}
	if cfg.Output == "" {
		slog.Error("Judge mode requires --output")
		os.Exit(1)
	}

	pf, err := pool.ReadPoolFile(cfg.PoolPath)
	if err != nil {
		slog.Error("Failed to read pool file", "error", err)
		os.Exit(1)
	}

	if err := judgment.ExportForAnnotation(pf, cfg.Output); err != nil {
		slog.Error("Failed to export annotation template", "error", err)
		os.Exit(1)
	}
	slog.Info("Annotation template written", "path", cfg.Output)
}

func buildQuickSpec(cfg cliConfig) *spec.BenchSpec {
	engines := make(map[string]spec.Engine)
	var engineNames []string

	if cfg.PgConnStr != "" {
		engines["pg-native"] = spec.Engine{Type: "postgres", Connection: cfg.PgConnStr}
		engineNames = append(engineNames, "pg-native")
	}
	if cfg.EsAddresses != "" {
		engines["elasticsearch"] = spec.Engine{Type: "elasticsearch", Connection: cfg.EsAddresses, Index: cfg.EsIndex}
		engineNames = append(engineNames, "elasticsearch")
	}

	return &spec.BenchSpec{
		Engines: engines,
		Jobs: []spec.Job{
			{
				Name:    "quick",
				Suite:   cfg.SuitePath,
				Engines: engineNames,
				Layer:   cfg.Layer,
			},
		},
	}
}
