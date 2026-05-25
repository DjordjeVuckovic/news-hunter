package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/trackctx"
	"github.com/spf13/cobra"
)

type validateFlags struct {
	trackArg  string
	specPath  string
	suitePath string
	failFast  bool
}

func newValidateCmd() *cobra.Command {
	var f validateFlags
	cmd := &cobra.Command{
		Use:   "validate [track]",
		Short: "Dry-run every query through each engine and report broken ones",
		Long: `Validates spec + suite ahead of a real pool/run:

  - templates render with the params provided
  - postgres queries pass EXPLAIN (syntax, columns, operators)
  - elasticsearch queries pass _validate/query (JSON, fields, types)
  - api descriptors parse as {method, path, body?, params?}

Returns non-zero exit if any query fails — wire it into CI.`,
		Example: `  bench validate fts_quality
  bench validate --track tracks/fts_quality --fail-fast`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeValidate(cmd, f, args)
		},
	}
	cmd.Flags().StringVar(&f.trackArg, "track", "", "Track name or path")
	cmd.Flags().StringVar(&f.specPath, "spec", "", "Override spec.yaml path")
	cmd.Flags().StringVar(&f.suitePath, "suite", "", "Override suite.yaml path (all jobs share it)")
	cmd.Flags().BoolVar(&f.failFast, "fail-fast", false, "Stop at first failure")
	return cmd
}

type validateRow struct {
	queryID string
	engine  string
	status  string
	detail  string
}

func executeValidate(cmd *cobra.Command, f validateFlags, args []string) error {
	tr, err := trackctx.Resolve(trackctx.Inputs{
		TrackArg:  trackArg(f.trackArg, args),
		SpecPath:  f.specPath,
		SuitePath: f.suitePath,
	})
	if err != nil {
		return err
	}

	bs, err := spec.LoadFromFile(tr.Spec)
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	executors, cleanup, err := createExecutors(cmd.Context(), bs)
	if err != nil {
		return fmt.Errorf("create executors: %w", err)
	}
	defer cleanup()

	var rows []validateRow
	failures := 0
	suites := map[string]*suite.LoadedSuite{}
	// seen deduplicates (suitePath, queryID, engineName) triples — two jobs
	// that share the same suite and engines would otherwise re-validate the
	// same pairs, doubling traffic and output noise.
	seen := map[string]struct{}{}

	for _, job := range bs.Jobs {
		ls, ok := suites[job.Suite]
		if !ok {
			loaded, err := suite.LoadFromFile(job.Suite)
			if err != nil {
				return fmt.Errorf("load suite for job %q: %w", job.Name, err)
			}
			suites[job.Suite] = loaded
			ls = loaded
		}
		for _, q := range ls.Suite.Queries {
			for _, engName := range job.Engines {
				key := job.Suite + "\x00" + q.ID + "\x00" + engName
				if _, done := seen[key]; done {
					continue
				}
				seen[key] = struct{}{}

				row := validateRow{queryID: q.ID, engine: engName}
				row = validateOne(cmd.Context(), row, q, engName, ls, executors[engName])
				rows = append(rows, row)
				if row.status != "OK" && row.status != "SKIP" {
					failures++
					if f.failFast {
						printValidateRows(cmd.OutOrStdout(), rows)
						return fmt.Errorf("validation failed (fail-fast)")
					}
				}
			}
		}
	}

	printValidateRows(cmd.OutOrStdout(), rows)
	cmd.Printf("\nTotal: %d checks, %d failed\n", len(rows), failures)
	if failures > 0 {
		return fmt.Errorf("%d query/engine pair(s) failed validation", failures)
	}
	return nil
}

func validateOne(ctx context.Context, row validateRow, q suite.Query, engName string, ls *suite.LoadedSuite, exec engine.Executor) validateRow {
	resolved, err := q.ResolveEngineQuery(engName, ls.Registry, ls.Dir)
	if err != nil {
		row.status = "TEMPLATE_ERR"
		row.detail = err.Error()
		return row
	}
	if resolved == nil {
		row.status = "SKIP"
		row.detail = "no query for this engine"
		return row
	}
	v, ok := exec.(engine.Validator)
	if !ok {
		row.status = "UNSUPPORTED"
		row.detail = "executor does not implement Validator"
		return row
	}
	if err := v.Validate(ctx, resolved.Query); err != nil {
		row.status = "INVALID"
		row.detail = truncate(err.Error(), 120)
		return row
	}
	row.status = "OK"
	return row
}

func printValidateRows(w io.Writer, rows []validateRow) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "QUERY\tENGINE\tSTATUS\tDETAIL")
	fmt.Fprintln(tw, "-----\t------\t------\t------")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.queryID, r.engine, r.status, r.detail)
	}
	tw.Flush()
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
