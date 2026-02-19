package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

func WriteTable(r *Report, w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "\n=== FTS Quality Benchmark ===\n")

	for _, jr := range r.Jobs {
		fmt.Fprintf(tw, "\n--- Job: %s ---\n\n", jr.JobName)
		writeAggregatedTable(tw, &jr, r.Config.KValues)
		writeLatencyTable(tw, &jr)
		writePerQueryTable(tw, &jr, r.Config.KValues)
	}

	tw.Flush()
}

func writeAggregatedTable(tw *tabwriter.Writer, jr *JobReport, kValues []int) {
	fmt.Fprintf(tw, "Aggregated Results (mean across %d queries)\n\n", countQueries(jr))

	header := []string{"Engine"}
	for _, k := range kValues {
		header = append(header, fmt.Sprintf("NDCG@%d", k))
	}
	for _, k := range kValues {
		header = append(header, fmt.Sprintf("P@%d", k))
	}
	header = append(header, "MAP", "MRR", "Errors")
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, agg := range jr.Aggregated {
		row := []string{agg.EngineName}
		for _, k := range kValues {
			row = append(row, fmt.Sprintf("%.4f", agg.NDCG[k]))
		}
		for _, k := range kValues {
			row = append(row, fmt.Sprintf("%.4f", agg.Precision[k]))
		}
		row = append(row,
			fmt.Sprintf("%.4f", agg.MAP),
			fmt.Sprintf("%.4f", agg.MRR),
			fmt.Sprintf("%d/%d", agg.ErrorCount, agg.QueryCount),
		)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	fmt.Fprintln(tw)
}

func writeLatencyTable(tw *tabwriter.Writer, jr *JobReport) {
	fmt.Fprintf(tw, "Latency Statistics (aggregated across queries)\n\n")

	header := []string{"Engine", "Min", "p50", "p75", "p90", "p95", "p99", "Max", "Mean", "Stddev", "Samples"}
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, agg := range jr.Aggregated {
		s := agg.Latency
		row := []string{
			agg.EngineName,
			fmtDuration(s.Min),
			fmtDuration(s.P50()),
			fmtDuration(s.P75()),
			fmtDuration(s.P90()),
			fmtDuration(s.P95()),
			fmtDuration(s.P99()),
			fmtDuration(s.Max),
			fmtDuration(s.Mean),
			fmtDuration(s.Stddev),
			fmt.Sprintf("%d", s.SampleCount),
		}
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	fmt.Fprintln(tw)
}

func writePerQueryTable(tw *tabwriter.Writer, jr *JobReport, kValues []int) {
	fmt.Fprintf(tw, "Per-Query Results\n\n")

	k := primaryK(kValues)

	header := []string{"Query", "Engine", fmt.Sprintf("NDCG@%d", k), fmt.Sprintf("P@%d", k), "AP", "RR", "Hits", "p50", "p95", "Status"}
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, e := range jr.PerQuery {
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
			fmtDuration(e.Latency.P50()),
			fmtDuration(e.Latency.P95()),
			status,
		}
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	fmt.Fprintln(tw)
}

func primaryK(kValues []int) int {
	if len(kValues) > 0 {
		return kValues[len(kValues)-1]
	}
	return 10
}

func countQueries(jr *JobReport) int {
	if len(jr.Aggregated) == 0 {
		return 0
	}
	return jr.Aggregated[0].QueryCount
}

func fmtScore(scores map[int]float64, k int) string {
	if scores == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.4f", scores[k])
}

func fmtDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
