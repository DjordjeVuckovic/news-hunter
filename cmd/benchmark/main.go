package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/benchmark/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/benchmark/report"
	"github.com/DjordjeVuckovic/news-hunter/internal/benchmark/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/benchmark/suite"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg/native"
)

func main() {
	cfg := parseFlags()

	ctx := context.Background()

	s, err := suite.LoadFromFile(cfg.SuitePath)
	if err != nil {
		slog.Error("Failed to load benchmark suite", "path", cfg.SuitePath, "error", err)
		os.Exit(1)
	}
	slog.Info("Loaded benchmark suite", "name", s.Name, "queries", len(s.Queries))

	engines, cleanup, err := createEngines(ctx, cfg)
	if err != nil {
		slog.Error("Failed to create search engines", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	if len(engines) == 0 {
		slog.Error("No engines configured. Use --pg and/or --es-addresses flags.")
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
	}

	r := runner.New(s, engines, runCfg)
	result, err := r.Run(ctx)
	if err != nil {
		slog.Error("Benchmark run failed", "error", err)
		os.Exit(1)
	}

	rpt := report.Generate(result)
	report.WriteTable(rpt, os.Stdout)
}

func createEngines(ctx context.Context, cfg cliConfig) ([]engine.SearchEngine, func(), error) {
	var engines []engine.SearchEngine
	var cleanups []func()

	if cfg.PgConnStr != "" {
		pool, err := native.NewConnectionPool(ctx, native.PoolConfig{ConnStr: cfg.PgConnStr})
		if err != nil {
			return nil, nil, fmt.Errorf("pg connection: %w", err)
		}
		cleanups = append(cleanups, pool.Close)

		searcher, err := native.NewReader(pool)
		if err != nil {
			pool.Close()
			return nil, nil, fmt.Errorf("pg searcher: %w", err)
		}

		engines = append(engines, engine.SearchEngine{Name: "pg-native", Searcher: searcher})
		slog.Info("Enabled engine", "name", "pg-native")
	}

	if cfg.EsAddresses != "" {
		searcher, err := es.NewSearcher(es.ClientConfig{
			Addresses: strings.Split(cfg.EsAddresses, ","),
			IndexName: cfg.EsIndex,
		})
		if err != nil {
			for _, c := range cleanups {
				c()
			}
			return nil, nil, fmt.Errorf("es searcher: %w", err)
		}

		engines = append(engines, engine.SearchEngine{Name: "elasticsearch", Searcher: searcher})
		slog.Info("Enabled engine", "name", "elasticsearch")
	}

	cleanup := func() {
		for _, c := range cleanups {
			c()
		}
	}

	return engines, cleanup, nil
}
