package main

import (
	"fmt"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/report"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/spf13/cobra"
)

type runFlags struct {
	specPath   string
	suite      string
	pg         string
	es         string
	esIndex    string
	api        string
	kValues    string
	maxK       int
	warmup     int
	iterations int
	output     string
}

func newRunCmd() *cobra.Command {
	var f runFlags
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run a benchmark with judgments wired in",
		Long:    "Executes every job in the spec, computes IR-quality metrics + latency stats, and writes the report.",
		Example: "  bench run --spec configs/bench/spec.yaml --output results/report.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeRun(cmd, f)
		},
	}
	cmd.Flags().StringVar(&f.specPath, "spec", "", "Path to bench spec YAML (multi-job mode)")
	cmd.Flags().StringVar(&f.suite, "suite", "configs/bench/fts_quality_v1.yaml", "Suite YAML (quick single-job mode when --spec is empty)")
	cmd.Flags().StringVar(&f.pg, "pg", "", "Postgres connection string (quick mode)")
	cmd.Flags().StringVar(&f.es, "es-addresses", "", "Elasticsearch base URL (quick mode)")
	cmd.Flags().StringVar(&f.esIndex, "es-index", "articles", "Elasticsearch index name (quick mode)")
	cmd.Flags().StringVar(&f.api, "api", "", "news-hunter API base URL (quick mode)")
	cmd.Flags().StringVar(&f.kValues, "k", "3,5,10", "K cut-offs for NDCG/P/R/F1, comma-separated")
	cmd.Flags().IntVar(&f.maxK, "max-k", 100, "Max docs retrieved per query")
	cmd.Flags().IntVar(&f.warmup, "warmup", 0, "Warmup iterations before measurement")
	cmd.Flags().IntVar(&f.iterations, "iterations", 1, "Measured iterations (median latency used)")
	cmd.Flags().StringVar(&f.output, "output", "", "Optional path to write JSON report")
	return cmd
}

func executeRun(cmd *cobra.Command, f runFlags) error {
	ks, err := parseKList(f.kValues)
	if err != nil {
		return err
	}

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
		KValues:            ks,
		MaxK:               f.maxK,
		RelevanceThreshold: runner.DefaultRelevanceThreshold,
		WarmupRuns:         f.warmup,
		Runs:               max(f.iterations, 1),
	}
	if f.specPath != "" {
		applySpecRunOverrides(&runCfg, bs)
	}

	executors, cleanup, err := createExecutors(cmd.Context(), bs)
	if err != nil {
		return fmt.Errorf("create executors: %w", err)
	}
	defer cleanup()

	r := runner.New(runCfg)
	result, err := r.RunAll(cmd.Context(), bs, executors)
	if err != nil {
		return fmt.Errorf("run benchmark: %w", err)
	}

	rpt := report.Generate(result, &report.GenerateOptions{Spec: bs})
	report.WriteTable(rpt, os.Stdout)

	if f.output != "" {
		if err := report.WriteJSON(rpt, f.output); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
		cmd.Printf("Report written: %s\n", f.output)
	}
	return nil
}

func applySpecRunOverrides(cfg *runner.Config, bs *spec.BenchSpec) {
	if bs.Runs.Warmup > 0 && cfg.WarmupRuns == 0 {
		cfg.WarmupRuns = bs.Runs.Warmup
	}
	if bs.Runs.Iterations > 0 && cfg.Runs <= 1 {
		cfg.Runs = bs.Runs.Iterations
	}
	if len(bs.Metrics.KValues) > 0 {
		cfg.KValues = bs.Metrics.KValues
	}
	if bs.Metrics.MaxK > 0 {
		cfg.MaxK = bs.Metrics.MaxK
	}
	if bs.Metrics.RelevanceThreshold > 0 {
		cfg.RelevanceThreshold = bs.Metrics.RelevanceThreshold
	}
}
