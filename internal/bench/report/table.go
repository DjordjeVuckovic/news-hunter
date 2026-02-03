package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func WriteTable(r *Report, w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "\n=== FTS Quality Benchmark ===\n\n")

	writeAggregatedTable(tw, r)
	writePerQueryTable(tw, r)

	tw.Flush()
}

func writeAggregatedTable(tw *tabwriter.Writer, r *Report) {
	fmt.Fprintf(tw, "--- Aggregated Results (mean across %d queries) ---\n\n", countQueries(r))

	header := []string{"Engine"}
	for _, k := range r.Config.KValues {
		header = append(header, fmt.Sprintf("NDCG@%d", k))
	}
	for _, k := range r.Config.KValues {
		header = append(header, fmt.Sprintf("P@%d", k))
	}
	header = append(header, "MAP", "MRR", "Latency", "Errors")
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, agg := range r.Aggregated {
		row := []string{agg.EngineName}
		for _, k := range r.Config.KValues {
			row = append(row, fmt.Sprintf("%.4f", agg.NDCG[k]))
		}
		for _, k := range r.Config.KValues {
			row = append(row, fmt.Sprintf("%.4f", agg.Precision[k]))
		}
		row = append(row,
			fmt.Sprintf("%.4f", agg.MAP),
			fmt.Sprintf("%.4f", agg.MRR),
			agg.MeanLatency.Truncate(100).String(),
			fmt.Sprintf("%d/%d", agg.ErrorCount, agg.QueryCount),
		)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	fmt.Fprintln(tw)
}

func writePerQueryTable(tw *tabwriter.Writer, r *Report) {
	fmt.Fprintf(tw, "--- Per-Query Results ---\n\n")

	k := primaryK(r)

	header := []string{"Query", "Engine", fmt.Sprintf("NDCG@%d", k), fmt.Sprintf("P@%d", k), "AP", "RR", "Hits", "Latency", "Status"}
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, e := range r.PerQuery {
		status := "OK"
		if e.Error != "" {
			status = "ERR"
		}
		row := []string{
			e.QueryID,
			e.EngineName,
			fmtScore(e.NDCG, k),
			fmtScore(e.Precision, k),
			fmt.Sprintf("%.4f", e.AP),
			fmt.Sprintf("%.4f", e.RR),
			fmt.Sprintf("%d", e.TotalMatches),
			e.Latency.Truncate(100).String(),
			status,
		}
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	fmt.Fprintln(tw)
}

func primaryK(r *Report) int {
	if len(r.Config.KValues) > 0 {
		return r.Config.KValues[len(r.Config.KValues)-1]
	}
	return 10
}

func countQueries(r *Report) int {
	if len(r.Aggregated) == 0 {
		return 0
	}
	return r.Aggregated[0].QueryCount
}

func fmtScore(scores map[int]float64, k int) string {
	if scores == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.4f", scores[k])
}
