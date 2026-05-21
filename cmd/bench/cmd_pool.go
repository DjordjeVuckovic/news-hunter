package main

import (
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
	"github.com/spf13/cobra"
)

type poolFlags struct {
	specPath string
	suite    string
	pg       string
	es       string
	esIndex  string
	api      string
	depth    int
	output   string
}

func newPoolCmd() *cobra.Command {
	var f poolFlags
	cmd := &cobra.Command{
		Use:     "pool",
		Short:   "Run queries and collect a TREC-style candidate pool",
		Long:    "Executes the suite against every engine and writes a deduplicated pool of doc IDs per query for downstream judging.",
		Example: "  bench pool --spec configs/bench/spec.yaml --output configs/bench/trec/pool_v1.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executePool(cmd, f)
		},
	}
	cmd.Flags().StringVar(&f.specPath, "spec", "", "Path to bench spec YAML")
	cmd.Flags().StringVar(&f.suite, "suite", "configs/bench/fts_quality_v1.yaml", "Suite YAML (quick mode)")
	cmd.Flags().StringVar(&f.pg, "pg", "", "Postgres connection string (quick mode)")
	cmd.Flags().StringVar(&f.es, "es-addresses", "", "Elasticsearch base URL (quick mode)")
	cmd.Flags().StringVar(&f.esIndex, "es-index", "articles", "Elasticsearch index (quick mode)")
	cmd.Flags().StringVar(&f.api, "api", "", "API base URL (quick mode)")
	cmd.Flags().IntVar(&f.depth, "depth", 100, "Top-K per engine to pool")
	cmd.Flags().StringVar(&f.output, "output", "", "Output pool YAML path (required)")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func executePool(cmd *cobra.Command, f poolFlags) error {
	bs, err := loadBenchSpec(f.specPath, quickSpecFlags{
		suitePath:   f.suite,
		pgConnStr:   f.pg,
		esAddresses: f.es,
		esIndex:     f.esIndex,
		apiBaseURL:  f.api,
	})
	if err != nil {
		return err
	}

	runCfg := runner.Config{
		KValues:            []int{f.depth},
		MaxK:               f.depth,
		RelevanceThreshold: runner.DefaultRelevanceThreshold,
		Runs:               1,
	}

	executors, cleanup, err := createExecutors(cmd.Context(), bs)
	if err != nil {
		return fmt.Errorf("create executors: %w", err)
	}
	defer cleanup()

	r := runner.New(runCfg)
	result, err := r.RunAll(cmd.Context(), bs, executors)
	if err != nil {
		return fmt.Errorf("pool run: %w", err)
	}

	descs, err := collectQueryDescriptions(bs)
	if err != nil {
		return fmt.Errorf("load query descriptions: %w", err)
	}

	pf := buildPoolFile(result, descs, f.depth)
	if err := pool.WritePoolFile(pf, f.output); err != nil {
		return fmt.Errorf("write pool: %w", err)
	}
	cmd.Printf("Pool written: %s (queries=%d)\n", f.output, len(pf.Queries))
	return nil
}

func buildPoolFile(result *runner.BenchmarkResult, descs map[string]string, depth int) *pool.PoolFile {
	pf := &pool.PoolFile{}
	for _, jr := range result.Jobs {
		pf.SuiteName = jr.JobName
		for _, qID := range jr.QueryOrder {
			engResults := jr.Results[qID]
			executions := make(map[string]*engine.Execution)
			for _, engName := range jr.EngineNames {
				qr, ok := engResults[engName]
				if !ok || qr.Error != nil {
					continue
				}
				executions[engName] = &engine.Execution{
					RankedDocIDs: qr.RankedDocIDs,
					TotalMatches: qr.TotalMatches,
				}
			}
			pf.Queries = append(pf.Queries, pool.PoolEntry{
				QueryID:   qID,
				QueryDesc: descs[qID],
				Docs:      pool.PoolResults(executions, depth),
			})
		}
	}
	return pf
}

func collectQueryDescriptions(bs *spec.BenchSpec) (map[string]string, error) {
	descs := make(map[string]string)
	seen := make(map[string]struct{})
	for _, job := range bs.Jobs {
		if _, ok := seen[job.Suite]; ok {
			continue
		}
		seen[job.Suite] = struct{}{}
		ls, err := suite.LoadFromFile(job.Suite)
		if err != nil {
			return nil, fmt.Errorf("load suite %q for job %q: %w", job.Suite, job.Name, err)
		}
		for _, q := range ls.Suite.Queries {
			if q.Description != "" {
				descs[q.ID] = q.Description
			}
		}
	}
	return descs, nil
}
