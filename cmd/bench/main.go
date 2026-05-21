package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

const (
	cliName  = "bench"
	cliShort = "Search engine quality + latency benchmark"
	cliLong  = `bench evaluates full-text, vector, and hybrid search queries against multiple
engines (Postgres, ParadeDB, Elasticsearch, the news-hunter API), produces
IR-quality metrics (NDCG, MAP, MRR, Bpref, P/R/F1) and latency statistics.

Typical pipeline:

  1. pool   — run all queries through every engine, gather candidate docs
  2. judge  — grade each pooled doc (keyword baseline, claude-cli, claude-api, stub)
  3. run    — execute the suite with judgments wired in, produce the report
  4. qrels  — (optional) export TREC qrels for external validation
`
)

func main() {
	root := &cobra.Command{
		Use:           cliName,
		Short:         cliShort,
		Long:          cliLong,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newRunCmd(),
		newPoolCmd(),
		newJudgeCmd(),
		newQrelsCmd(),
		newValidateCmd(),
		newInitCmd(),
		newShowCmd(),
	)

	if err := root.Execute(); err != nil {
		slog.Error("bench failed", "error", err)
		os.Exit(1)
	}
}
