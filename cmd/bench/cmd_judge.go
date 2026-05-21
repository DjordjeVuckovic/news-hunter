package main

import (
	"fmt"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/pool"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
	"github.com/spf13/cobra"
)

type judgeFlags struct {
	poolPath    string
	output      string
	strategy    string
	pg          string
	concurrency int
	batchSize   int
	resume      bool
	apiKey      string
	apiModel    string
	apiBaseURL  string
	cliBinary   string
}

func newJudgeCmd() *cobra.Command {
	var f judgeFlags
	cmd := &cobra.Command{
		Use:   "judge",
		Short: "Grade a pool file with the chosen strategy",
		Long: `Grades every (query, doc) pair in a pool file using one of:

  keyword     — deterministic token-overlap baseline (no network, no LLM)
  claude-cli  — invokes 'claude -p' per batch (Anthropic LLM-as-judge batched)
  claude-api  — Anthropic Messages API in batches (set ANTHROPIC_API_KEY)
  stub        — writes grade:-1 placeholders for manual human filling

Articles are fetched in batch per query and never leave memory — no large
intermediate "enriched pool" file is produced.

LLM strategies grade N candidates per call (Anthropic's "judge N candidates"
cookbook pattern) so wall-clock time drops ~10x vs one-by-one. Partial batch
responses are auto-retried per-doc.

The judge is resumable: re-running with the same --output appends to the
existing file, skipping (query_id, doc_id) pairs already graded. Output is
written atomically after every query — safe to Ctrl-C.`,
		Example: `  bench judge --pool pool_v1.yaml --strategy keyword --output annotations_kw.yaml
  bench judge --pool pool_v1.yaml --strategy claude-api --batch 20 --output annotations_llm.yaml
  bench judge --pool pool_v1.yaml --strategy claude-api --output annotations_llm.yaml --resume`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeJudge(cmd, f)
		},
	}
	cmd.Flags().StringVar(&f.poolPath, "pool", "", "Path to pool YAML (required)")
	cmd.Flags().StringVar(&f.output, "output", "", "Output judgment YAML path (required)")
	cmd.Flags().StringVar(&f.strategy, "strategy", string(judgment.StrategyKeyword), "Judge strategy: keyword | claude-cli | claude-api | stub")
	cmd.Flags().StringVar(&f.pg, "pg", "", "Postgres connection string (or set PG_CONNECTION_STRING)")
	cmd.Flags().IntVar(&f.concurrency, "concurrency", 4, "Parallel Grade calls for per-doc strategies (1-32)")
	cmd.Flags().IntVar(&f.batchSize, "batch", 0, "Override LLM batch size (0 = strategy default: api=20, cli=10)")
	cmd.Flags().BoolVar(&f.resume, "resume", false, "Resume from existing --output file, skip already-graded docs")
	cmd.Flags().StringVar(&f.apiKey, "api-key", "", "Anthropic API key for claude-api (or set ANTHROPIC_API_KEY)")
	cmd.Flags().StringVar(&f.apiModel, "api-model", "", "Anthropic model id (default: haiku-4.5)")
	cmd.Flags().StringVar(&f.apiBaseURL, "api-base", "", "Anthropic API base URL (advanced)")
	cmd.Flags().StringVar(&f.cliBinary, "cli-binary", "", "claude CLI binary path (default: claude)")
	_ = cmd.MarkFlagRequired("pool")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func executeJudge(cmd *cobra.Command, f judgeFlags) error {
	pf, err := pool.ReadPoolFile(f.poolPath)
	if err != nil {
		return fmt.Errorf("read pool: %w", err)
	}

	kind := judgment.StrategyKind(f.strategy)

	if kind == judgment.StrategyStub {
		jf := buildStubJudgments(pf)
		if err := judgment.WriteFile(jf, f.output); err != nil {
			return fmt.Errorf("write judgments: %w", err)
		}
		cmd.Printf("Stub judgments written: %s (queries=%d)\n", f.output, len(jf.Queries))
		return nil
	}

	strat, err := judgment.NewStrategy(kind, judgment.StrategyOptions{
		APIKey:      envOrFlag("ANTHROPIC_API_KEY", f.apiKey),
		APIModel:    f.apiModel,
		APIBaseURL:  f.apiBaseURL,
		CLIBinary:   f.cliBinary,
		Concurrency: f.concurrency,
	})
	if err != nil {
		return err
	}

	conn := envOrFlag("PG_CONNECTION_STRING", f.pg)
	if conn == "" {
		return fmt.Errorf("judge requires --pg or PG_CONNECTION_STRING for article enrichment")
	}
	reader, err := factory.NewReader(cmd.Context(), factory.StorageConfig{
		Type: storage.PG,
		Pg:   &pg.PoolConfig{ConnStr: conn},
	})
	if err != nil {
		return fmt.Errorf("create reader: %w", err)
	}

	writer := judgment.NewIncrementalWriter(f.output, strat.Name())
	var prior *judgment.JudgmentFile
	if f.resume {
		prior, err = writer.LoadPrior()
		if err != nil {
			return fmt.Errorf("load prior judgments: %w", err)
		}
		if prior != nil {
			cmd.Printf("Resume: loaded %d prior queries from %s\n", len(prior.Queries), f.output)
		}
	}

	jrunner := judgment.NewRunner(judgment.RunnerConfig{
		Strategy:    strat,
		Reader:      reader,
		Concurrency: f.concurrency,
		BatchSize:   f.batchSize,
		Existing:    prior,
		Sink:        writer.Append,
		OnQueryStart: func(qid string, total, skipped int) {
			if skipped > 0 {
				cmd.Printf("[%s] grading %d docs (%d already done, skipping)\n", qid, total-skipped, skipped)
			} else {
				cmd.Printf("[%s] grading %d docs\n", qid, total)
			}
		},
		OnBatch: func(bp judgment.BatchProgress) {
			cmd.Printf("  └ batch %d/%d: graded=%d missing=%d %s\n",
				bp.BatchIdx, bp.BatchN, bp.Graded, bp.Missing, formatHistogram(bp.Histogram))
		},
		OnQueryDone: func(qp judgment.QueryProgress) {
			cmd.Printf("[%s] done: graded=%d skipped=%d unjudged=%d %s\n",
				qp.QueryID, qp.Graded, qp.Skipped, qp.Unjudged, formatHistogram(qp.Histogram))
		},
	})

	if _, err := jrunner.Run(cmd.Context(), pf); err != nil {
		return fmt.Errorf("judge run: %w", err)
	}

	final := writer.Snapshot()
	cmd.Printf("Judgments written: %s (strategy=%s, queries=%d)\n", f.output, final.Strategy, len(final.Queries))
	return nil
}

func buildStubJudgments(pf *pool.PoolFile) *judgment.JudgmentFile {
	jf := &judgment.JudgmentFile{
		Strategy: string(judgment.StrategyStub),
		Queries:  make([]judgment.JudgmentEntry, 0, len(pf.Queries)),
	}
	for _, entry := range pf.Queries {
		docs := make([]judgment.GradedDoc, 0, len(entry.Docs))
		for _, d := range entry.Docs {
			docs = append(docs, judgment.GradedDoc{DocID: d.DocID, Grade: judgment.GradeUnjudged})
		}
		jf.Queries = append(jf.Queries, judgment.JudgmentEntry{QueryID: entry.QueryID, Docs: docs})
	}
	return jf
}

func formatHistogram(h map[int]int) string {
	if len(h) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("(")
	for g := 3; g >= 0; g-- {
		if g < 3 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%d:%d", g, h[g])
	}
	b.WriteString(")")
	return b.String()
}
