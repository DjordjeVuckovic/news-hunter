package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

// table-level color instances — only applied to last columns or standalone
// lines to avoid ANSI-byte misalignment inside tabwriter.
var (
	tBold   = color.New(color.Bold)
	tGreen  = color.New(color.FgGreen, color.Bold)
	tRed    = color.New(color.FgRed, color.Bold)
	tYellow = color.New(color.FgYellow)
	tDim    = color.New(color.FgHiBlack)
)

func WriteTable(r *Report, w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	title := r.Provenance.SpecID
	if title == "" {
		title = "Benchmark"
	}
	fmt.Fprintf(tw, "\n%s\n", tBold.Sprintf("=== %s  run_id=%s ===", title, r.Provenance.RunID))

	for _, jr := range r.Jobs {
		fmt.Fprintf(tw, "\n%s\n\n", tBold.Sprintf("--- Job: %s ---", jr.JobName))
		if !hasAnyJudgments(&jr) {
			fmt.Fprintf(tw, "%s No relevance judgments found. Showing latency only.\n", tYellow.Sprint("WARNING:"))
			fmt.Fprintf(tw, "         Run bench pool first, then bench judge.\n\n")
			writeLatencyTable(tw, &jr)
		} else {
			writeAggregatedTable(tw, &jr, r.Config.KValues)
			writeLatencyTable(tw, &jr)
			writeSignificanceTable(tw, &jr)
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
	fmt.Fprintf(tw, "%s\n\n", tBold.Sprint("Aggregated Results (mean across judged queries)"))

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
	fmt.Fprintf(tw, "%s\n\n", tBold.Sprint("Latency Statistics (aggregated across queries)"))

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

func writeSignificanceTable(tw *tabwriter.Writer, jr *JobReport) {
	if len(jr.Significance) == 0 {
		return
	}
	fmt.Fprintf(tw, "%s\n\n", tBold.Sprint("Statistical Significance (Wilcoxon signed-rank, two-tailed; * p<0.05, ** p<0.01)"))
	fmt.Fprintln(tw, "Engine A\tEngine B\tMetric\tW\tp-value\tSig")
	fmt.Fprintln(tw, "---\t---\t---\t---\t---\t---")
	for _, s := range jr.Significance {
		// Sig is the last column — safe to add ANSI without breaking alignment.
		sig := colorSig(s.Stars)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%.1f\t%.4f\t%s\n",
			s.EngineA, s.EngineB, s.Metric, s.W, s.P, sig)
	}
	fmt.Fprintln(tw)
}

func colorSig(stars string) string {
	switch stars {
	case "**":
		return tGreen.Sprint("**")
	case "*":
		return tYellow.Sprint("*")
	default:
		return tDim.Sprint("ns")
	}
}

func writePerQueryTable(tw *tabwriter.Writer, jr *JobReport, kValues []int) {
	fmt.Fprintf(tw, "%s\n\n", tBold.Sprint("Per-Query Results"))

	k := primaryK(kValues)

	header := []string{"Query", "Engine", fmt.Sprintf("NDCG@%d", k), fmt.Sprintf("P@%d", k), "AP", "RR", "Bpref", "Hits", "p50", "p95", "Status"}
	fmt.Fprintln(tw, strings.Join(header, "\t"))

	sep := make([]string, len(header))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))

	for _, e := range jr.PerQuery {
		// Status is the last column — safe to color without breaking tabwriter alignment.
		var status string
		if e.Error != "" {
			status = tRed.Sprint("ERR")
		} else {
			status = tGreen.Sprint("OK")
		}

		apStr, rrStr, bprefStr := tDim.Sprint("N/A"), tDim.Sprint("N/A"), tDim.Sprint("N/A")
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
