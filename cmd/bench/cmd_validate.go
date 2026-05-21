package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
	"github.com/spf13/cobra"
)

type validateFlags struct {
	specPath string
	suite    string
	pg       string
	es       string
	esIndex  string
	api      string
	failFast bool
}

func newValidateCmd() *cobra.Command {
	var f validateFlags
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Dry-run every query through each engine and report broken ones",
		Long: `Validates spec + suite ahead of a real run:

  - templates render with the params provided
  - postgres queries pass EXPLAIN (syntax, columns, operators)
  - elasticsearch queries pass _validate/query (JSON, fields, types)
  - api descriptors parse as {method, path, body?, params?}

Returns non-zero exit if any query fails — wire it into CI.`,
		Example: "  bench validate --spec configs/bench/spec.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeValidate(cmd, f)
		},
	}
	cmd.Flags().StringVar(&f.specPath, "spec", "", "Path to bench spec YAML")
	cmd.Flags().StringVar(&f.suite, "suite", "configs/bench/fts_quality_v1.yaml", "Suite YAML (quick mode)")
	cmd.Flags().StringVar(&f.pg, "pg", "", "Postgres connection (quick mode)")
	cmd.Flags().StringVar(&f.es, "es-addresses", "", "Elasticsearch base URL (quick mode)")
	cmd.Flags().StringVar(&f.esIndex, "es-index", "articles", "Elasticsearch index (quick mode)")
	cmd.Flags().StringVar(&f.api, "api", "", "API base URL (quick mode)")
	cmd.Flags().BoolVar(&f.failFast, "fail-fast", false, "Stop at first failure instead of reporting all")
	return cmd
}

type validateRow struct {
	queryID string
	engine  string
	status  string // OK | TEMPLATE_ERR | UNSUPPORTED | INVALID
	detail  string
}

func executeValidate(cmd *cobra.Command, f validateFlags) error {
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

	executors, cleanup, err := createExecutors(cmd.Context(), bs)
	if err != nil {
		return fmt.Errorf("create executors: %w", err)
	}
	defer cleanup()

	var rows []validateRow
	failures := 0

	// Multiple jobs often share the same suite. Load each path once.
	suites := map[string]*suite.LoadedSuite{}
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
				row := validateRow{queryID: q.ID, engine: engName}
				row = validateOne(cmd.Context(), row, q, engName, ls, executors[engName])
				rows = append(rows, row)
				if row.status != "OK" {
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
		// Engine not configured for this query — not a failure.
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

func printValidateRows(w interface{ Write(p []byte) (int, error) }, rows []validateRow) {
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

// Ensure os is used (some builds elide imports otherwise; kept explicit).
var _ = os.Stdout
