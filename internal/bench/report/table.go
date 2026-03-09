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
		if !hasAnyJudgments(&jr) {
			fmt.Fprintf(tw, "WARNING: No relevance judgments found. Showing latency only.\n")
			fmt.Fprintf(tw, "         Run pool mode first, annotate, then merge judgments into the suite.\n\n")
			writeLatencyTable(tw, &jr)
		} else {
			writeAggregatedTable(tw, &jr, r.Config.KValues)
			writeLatencyTable(tw, &jr)
			writePerQueryTable(tw, &jr, r.Config.KValues)
		}
	}

	tw.Flush()
}

func hasAnyJudgments(jr *JobReport) bool {
	for _, e := range jr.PerQuery {
		if e.Judged {
			return true
		}
	}
	return false
}

func writeAggregatedTable(tw *tabwriter.Writer, jr *JobReport, kValues []int) {
	fmt.Fprintf(tw, "Aggregated Results (mean across judged queries)\n\n")

	header := []string{"Engine"}
	for _, k := range kValues {
		header = append(header, fmt.Sprintf("NDCG@%d", k))
	}
	for _, k := range kValues {
		header = append(header, fmt.Sprintf("P@%d", k))
	}
	header = append(header, "MAP", "MRR", "Bpref", "Judged", "Errors")
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, agg := range jr.Aggregated {
		row := []string{agg.EngineName}
		if agg.JudgedCount > 0 {
			for _, k := range kValues {
				row = append(row, fmt.Sprintf("%.4f", agg.NDCG[k]))
			}
			for _, k := range kValues {
				row = append(row, fmt.Sprintf("%.4f", agg.Precision[k]))
			}
			row = append(row,
				fmt.Sprintf("%.4f", agg.MAP),
				fmt.Sprintf("%.4f", agg.MRR),
				fmt.Sprintf("%.4f", agg.MBpref),
			)
		} else {
			naCount := len(kValues)*2 + 3
			for i := 0; i < naCount; i++ {
				row = append(row, "N/A")
			}
		}
		row = append(row,
			fmt.Sprintf("%d/%d", agg.JudgedCount, agg.QueryCount),
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

	header := []string{"Query", "Engine", fmt.Sprintf("NDCG@%d", k), fmt.Sprintf("P@%d", k), "AP", "RR", "Bpref", "Hits", "p50", "p95", "Status"}
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

		apStr, rrStr, bprefStr := "N/A", "N/A", "N/A"
		if e.Judged {
			apStr = fmt.Sprintf("%.4f", e.AP)
			rrStr = fmt.Sprintf("%.4f", e.RR)
			bprefStr = fmt.Sprintf("%.4f", e.Bpref)
		}

		row := []string{
			e.QueryID,
			e.EngineName,
			fmtScore(e.NDCG, k),
			fmtScore(e.Precision, k),
			apStr,
			rrStr,
			bprefStr,
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
