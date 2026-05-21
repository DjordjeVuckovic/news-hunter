package main

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Inspect bench artifacts (pool, judgments, spec)",
		Long: `Pretty-prints a one-page summary of a bench artifact: query counts, grade
histograms, engine coverage, dedup ratios. The single best way to sanity-check
intermediates without grepping YAML.`,
	}
	cmd.AddCommand(newShowPoolCmd(), newShowJudgmentsCmd(), newShowSpecCmd())
	return cmd
}

func newShowPoolCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "pool <pool.yaml>",
		Short:   "Summarise a pool file",
		Args:    cobra.ExactArgs(1),
		Example: "  bench show pool configs/bench/trec/pool_v1.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := pool.ReadPoolFile(args[0])
			if err != nil {
				return err
			}
			showPool(cmd.OutOrStdout(), pf)
			return nil
		},
	}
}

func newShowJudgmentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "judgments <annotations.yaml>",
		Short:   "Summarise a judgments file",
		Args:    cobra.ExactArgs(1),
		Example: "  bench show judgments configs/bench/trec/annotations_v1.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			jf, err := judgment.ReadFile(args[0])
			if err != nil {
				return err
			}
			showJudgments(cmd.OutOrStdout(), jf)
			return nil
		},
	}
}

func newShowSpecCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "spec <spec.yaml>",
		Short:   "Summarise a bench spec",
		Args:    cobra.ExactArgs(1),
		Example: "  bench show spec configs/bench/spec.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			bs, err := spec.LoadFromFile(args[0])
			if err != nil {
				return err
			}
			showSpec(cmd.OutOrStdout(), bs)
			return nil
		},
	}
}

func showPool(w io.Writer, pf *pool.PoolFile) {
	fmt.Fprintf(w, "Pool: %s\n", pf.SuiteName)
	fmt.Fprintf(w, "Queries: %d\n\n", len(pf.Queries))

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "QUERY\tDOCS\tSOURCES\tDEDUP")
	fmt.Fprintln(tw, "-----\t----\t-------\t-----")

	var totalDocs, totalSourceHits int
	engineCounts := map[string]int{}

	for _, e := range pf.Queries {
		sources := map[string]int{}
		sourceHits := 0
		for _, d := range e.Docs {
			for _, s := range d.Sources {
				sources[s]++
				sourceHits++
				engineCounts[s]++
			}
		}
		totalDocs += len(e.Docs)
		totalSourceHits += sourceHits

		dedup := "—"
		if sourceHits > 0 {
			dedup = fmt.Sprintf("%.0f%%", 100*float64(sourceHits-len(e.Docs))/float64(sourceHits))
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", e.QueryID, len(e.Docs), formatSources(sources), dedup)
	}
	tw.Flush()

	fmt.Fprintf(w, "\nTotal unique docs: %d\n", totalDocs)
	fmt.Fprintf(w, "Total engine hits: %d\n", totalSourceHits)
	fmt.Fprintln(w, "\nPer-engine contribution:")
	for _, e := range sortedKeys(engineCounts) {
		fmt.Fprintf(w, "  %s: %d docs\n", e, engineCounts[e])
	}
}

func showJudgments(w io.Writer, jf *judgment.JudgmentFile) {
	fmt.Fprintf(w, "Judgments (strategy=%s)\n", jf.Strategy)
	fmt.Fprintf(w, "Queries: %d\n\n", len(jf.Queries))

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "QUERY\tTOTAL\tGRADED\tUNJUDGED\t3 (HI)\t2 (REL)\t1 (MARG)\t0 (NO)")
	fmt.Fprintln(tw, "-----\t-----\t------\t--------\t------\t-------\t--------\t------")

	totals := map[int]int{}
	allTotal, allGraded, allUnjudged := 0, 0, 0

	for _, qe := range jf.Queries {
		h := map[int]int{}
		graded, unjudged := 0, 0
		for _, d := range qe.Docs {
			if d.Grade < 0 {
				unjudged++
				continue
			}
			graded++
			h[d.Grade]++
			totals[d.Grade]++
		}
		allTotal += len(qe.Docs)
		allGraded += graded
		allUnjudged += unjudged
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
			qe.QueryID, len(qe.Docs), graded, unjudged, h[3], h[2], h[1], h[0])
	}
	tw.Flush()

	fmt.Fprintf(w, "\nTotal: %d docs across %d queries\n", allTotal, len(jf.Queries))
	fmt.Fprintf(w, "Graded: %d  Unjudged: %d\n", allGraded, allUnjudged)
	fmt.Fprintf(w, "Distribution: 3=%d  2=%d  1=%d  0=%d\n", totals[3], totals[2], totals[1], totals[0])
}

func showSpec(w io.Writer, bs *spec.BenchSpec) {
	fmt.Fprintln(w, "Engines:")
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "  NAME\tTYPE\tCONNECTION\tINDEX")
	for _, name := range sortedSpecKeys(bs.Engines) {
		e := bs.Engines[name]
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", name, e.Type, maskConn(e.Connection), e.Index)
	}
	tw.Flush()

	fmt.Fprintln(w, "\nJobs:")
	for _, j := range bs.Jobs {
		fmt.Fprintf(w, "  %s\n    suite:   %s\n    engines: %v\n", j.Name, j.Suite, j.Engines)
	}

	fmt.Fprintf(w, "\nMetrics: k=%v max_k=%d threshold=%d\n",
		bs.Metrics.KValues, bs.Metrics.MaxK, bs.Metrics.RelevanceThreshold)
	fmt.Fprintf(w, "Runs: warmup=%d iterations=%d\n", bs.Runs.Warmup, bs.Runs.Iterations)
}

func formatSources(sources map[string]int) string {
	if len(sources) == 0 {
		return "—"
	}
	keys := sortedKeys(sources)
	out := ""
	for i, k := range keys {
		if i > 0 {
			out += " "
		}
		out += fmt.Sprintf("%s:%d", k, sources[k])
	}
	return out
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedSpecKeys(m map[string]spec.Engine) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func maskConn(s string) string {
	if len(s) > 60 {
		return s[:30] + "…" + s[len(s)-15:]
	}
	return s
}
