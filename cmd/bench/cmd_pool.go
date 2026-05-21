package main

import (
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/meta"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/trackctx"
	"github.com/spf13/cobra"
)

type poolFlags struct {
	trackArg string
	specPath string
	output   string
	depth    int
}

func newPoolCmd() *cobra.Command {
	var f poolFlags
	cmd := &cobra.Command{
		Use:   "pool [track]",
		Short: "Run queries through all engines, write a TREC-style pool",
		Long: `Generates a deduplicated pool of candidate docs per query, ready to be
judged. Output goes to tracks/<name>/trec/pool.yaml by default; override
with --output for ad-hoc files.

The pool file carries a meta block (run_id, tool, engines, depth) so later
artifacts can attest which pool they were derived from.`,
		Example: `  bench pool fts_quality
  bench pool fts_quality --depth 50
  bench pool --track tracks/fts_quality --output /tmp/adhoc-pool.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executePool(cmd, f, args)
		},
	}
	cmd.Flags().StringVar(&f.trackArg, "track", "", "Track name or path")
	cmd.Flags().StringVar(&f.specPath, "spec", "", "Override spec.yaml path")
	cmd.Flags().IntVar(&f.depth, "depth", 0, "Top-K per engine (0 = spec.defaults.pool_depth or 100)")
	cmd.Flags().StringVar(&f.output, "output", "", "Override pool output path")
	return cmd
}

func executePool(cmd *cobra.Command, f poolFlags, args []string) error {
	tr, err := trackctx.Resolve(trackctx.Inputs{
		TrackArg:   trackArg(f.trackArg, args),
		SpecPath:   f.specPath,
		OutputPath: f.output,
	})
	if err != nil {
		return err
	}

	bs, err := spec.LoadFromFile(tr.Spec)
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	depth := f.depth
	if depth == 0 {
		depth = bs.Defaults.PoolDepth
	}
	if depth == 0 {
		depth = 100
	}

	runCfg := runner.Config{
		KValues: []int{depth},
		MaxK:    depth,
		Runs:    1,
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

	pf := buildPoolFile(result, descs, depth)
	pf.Meta = meta.New("pool")
	pf.Meta.SpecID = bs.ID
	pf.Meta.PoolDepth = depth
	pf.Meta.Engines = collectEngines(bs, result)

	outPath := f.output
	if outPath == "" {
		outPath = tr.Pool
	}
	if err := pool.WritePoolFile(pf, outPath); err != nil {
		return fmt.Errorf("write pool: %w", err)
	}
	cmd.Printf("Pool written: %s (queries=%d, run_id=%s)\n", outPath, len(pf.Queries), pf.Meta.RunID)
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

func collectEngines(bs *spec.BenchSpec, _ *runner.BenchmarkResult) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, job := range bs.Jobs {
		for _, eng := range job.Engines {
			if _, ok := seen[eng]; ok {
				continue
			}
			seen[eng] = struct{}{}
			out = append(out, eng)
		}
	}
	return out
}
