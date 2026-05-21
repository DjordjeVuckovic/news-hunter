package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/meta"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/report"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/trackctx"
	"github.com/spf13/cobra"
)

type runFlags struct {
	trackArg  string
	specPath  string
	suitePath string
	judgments string
	output    string
	kValues   string
	maxK      int
	warmup    int
	iters     int
}

func newRunCmd() *cobra.Command {
	var f runFlags
	cmd := &cobra.Command{
		Use:   "run [track]",
		Short: "Execute the benchmark + produce a report",
		Long: `Runs every job in the track's spec.yaml against its engines, computes IR
metrics + latency, writes a JSON report under tracks/<name>/reports/.

The judgments file used for scoring resolves in this order:
  1. --judgments <name|path>      (CLI override)
  2. spec.defaults.judgments      (per-track default)
  3. none → metrics-less report, warning printed`,
		Example: `  bench run fts_quality
  bench run fts_quality --judgments claude-api
  bench run --track tracks/fts_quality --judgments /tmp/ad-hoc-grades.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRun(cmd, f, args)
		},
	}
	cmd.Flags().StringVar(&f.trackArg, "track", "", "Track name (e.g. fts_quality) or path")
	cmd.Flags().StringVar(&f.specPath, "spec", "", "Override spec.yaml path")
	cmd.Flags().StringVar(&f.suitePath, "suite", "", "Override suite.yaml path")
	cmd.Flags().StringVar(&f.judgments, "judgments", "", "Strategy name (e.g. lexical) or annotations YAML path")
	cmd.Flags().StringVar(&f.output, "output", "", "Override report path (default: tracks/<name>/reports/<run_id>.json)")
	cmd.Flags().StringVar(&f.kValues, "k", "3,5,10", "K cut-offs for NDCG/P/R/F1")
	cmd.Flags().IntVar(&f.maxK, "max-k", 0, "Max docs retrieved per query (0 = spec.metrics.max_k)")
	cmd.Flags().IntVar(&f.warmup, "warmup", 0, "Warmup iterations")
	cmd.Flags().IntVar(&f.iters, "iterations", 0, "Measured iterations (0 = spec.runs.iterations)")
	return cmd
}

func executeRun(cmd *cobra.Command, f runFlags, args []string) error {
	ks, err := parseKList(f.kValues)
	if err != nil {
		return err
	}

	tr, err := trackctx.Resolve(trackctx.Inputs{
		TrackArg:   trackArg(f.trackArg, args),
		SpecPath:   f.specPath,
		SuitePath:  f.suitePath,
		OutputPath: f.output,
	})
	if err != nil {
		return err
	}

	bs, err := spec.LoadFromFile(tr.Spec)
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	// Wire judgments: --judgments wins, else spec.defaults.judgments, else nothing.
	judgmentsValue := f.judgments
	if judgmentsValue == "" {
		judgmentsValue = bs.Defaults.Judgments
	}
	jPath := tr.JudgmentsPath(judgmentsValue)

	judgmentsMap, err := loadJudgmentsMap(jPath)
	if err != nil {
		return err
	}

	runCfg := runner.Config{
		KValues:            ks,
		MaxK:               firstNonZero(f.maxK, bs.Metrics.MaxK),
		RelevanceThreshold: bs.Metrics.RelevanceThreshold,
		WarmupRuns:         firstNonZero(f.warmup, bs.Runs.Warmup),
		Runs:               max(firstNonZero(f.iters, bs.Runs.Iterations), 1),
		Judgments:          judgmentsMap,
	}
	if len(bs.Metrics.KValues) > 0 && !cmd.Flags().Changed("k") {
		runCfg.KValues = bs.Metrics.KValues
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
	rpt.Provenance.SpecID = bs.ID
	rpt.Provenance.Sources = &meta.Sources{
		Spec:      tr.Spec,
		Suite:     tr.Suite,
		Pool:      tr.Pool,
		Judgments: jPath,
	}

	// Print the table to stdout.
	report.WriteTable(rpt, os.Stdout)

	outPath := f.output
	if outPath == "" {
		outPath = tr.ReportPath(rpt.Provenance.RunID)
	}
	if err := report.WriteJSON(rpt, outPath); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	if outPath != "" && f.output == "" {
		if err := updateLatestPointer(tr.LatestReportPath(), outPath); err != nil {
			cmd.Printf("warning: could not update latest.json pointer: %v\n", err)
		}
	}
	cmd.Printf("Report written: %s\n", outPath)
	return nil
}

// loadJudgmentsMap reads a judgments YAML and flattens to a runner-friendly
// map[queryID][docID]grade. Returns nil if the path is empty or missing — the
// runner treats nil as "no judgments", and the report prints the appropriate
// warning. Filters out unjudged entries (grade < 0).
func loadJudgmentsMap(path string) (map[string]map[string]int, error) {
	if path == "" {
		return nil, nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	jf, err := judgment.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load judgments %s: %w", path, err)
	}
	out := make(map[string]map[string]int, len(jf.Queries))
	for _, qe := range jf.Queries {
		inner := make(map[string]int, len(qe.Docs))
		for _, d := range qe.Docs {
			if d.Grade < 0 {
				continue
			}
			inner[d.DocID.String()] = d.Grade
		}
		out[qe.QueryID] = inner
	}
	return out, nil
}

func firstNonZero(a, b int) int {
	if a != 0 {
		return a
	}
	return b
}

// updateLatestPointer writes a small JSON file ({"path":"reports/x.json"})
// next to the latest report. Cheap and doesn't depend on symlinks (which break
// on Windows / some FUSE mounts).
func updateLatestPointer(latestPath, actualPath string) error {
	rel, err := filepath.Rel(filepath.Dir(latestPath), actualPath)
	if err != nil {
		rel = actualPath
	}
	data := []byte(fmt.Sprintf(`{"latest": %q}`+"\n", rel))
	return os.WriteFile(latestPath, data, 0644)
}
