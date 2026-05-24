package main

import (
	"log/slog"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/spf13/cobra"
)

func init() {
	// Let the spec loader query the strategy registry without importing
	// judgment directly (keeps spec package dep-free of judgment).
	spec.KnownStrategies = judgment.KnownStrategies
}

const (
	cliName  = "bench"
	cliShort = "Search engine quality + latency benchmark"
	cliLong  = `bench evaluates full-text, vector, and hybrid search queries against multiple
engines (Postgres, ParadeDB, Elasticsearch, the news-hunter API), produces
IR-quality metrics (NDCG, MAP, MRR, Bpref, P/R/F1) and latency statistics.

Typical pipeline (track-first; pass the track name or path as a positional arg):

  1. init <name>      — scaffold tracks/<name>/{spec,suite}.yaml + folders
  2. validate <name>  — dry-run every query through each engine
  3. pool <name>      — gather candidate docs into tracks/<name>/trec/pool.yaml
  4. judge <name> --strategy lexical|claude-cli|claude-api|manual
                      — produces annotations.<strategy>.yaml
  5. run <name>       — execute the suite + report (auto-loads judgments per spec.defaults)
  6. qrels <name>     — (optional) export TREC qrels for trec_eval
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
