package main

import (
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/trackctx"
	"github.com/spf13/cobra"
)

type qrelsFlags struct {
	trackArg      string
	strategy      string
	judgmentsPath string
	output        string
}

func newQrelsCmd() *cobra.Command {
	var f qrelsFlags
	cmd := &cobra.Command{
		Use:   "qrels [track]",
		Short: "Export judgments to TREC qrels format",
		Long: `Converts an annotations YAML to standard TREC qrels (query_id 0 doc_id grade)
for use with trec_eval and external IR tooling.

Picks the annotations file in this order: --judgments PATH > --strategy NAME >
defaults to the track's lexical annotations.`,
		Example: `  bench qrels fts_quality --strategy lexical
  bench qrels fts_quality --strategy claude-api
  bench qrels --judgments /tmp/ad-hoc.yaml --output /tmp/q.tsv`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeQrels(cmd, f, args)
		},
	}
	cmd.Flags().StringVar(&f.trackArg, "track", "", "Track name or path")
	cmd.Flags().StringVar(&f.strategy, "strategy", "", "Strategy name (e.g. lexical) — resolves to trec/annotations.<name>.yaml")
	cmd.Flags().StringVar(&f.judgmentsPath, "judgments", "", "Override annotations YAML path")
	cmd.Flags().StringVar(&f.output, "output", "", "Override qrels TSV output path")
	return cmd
}

func executeQrels(cmd *cobra.Command, f qrelsFlags, args []string) error {
	tr, err := trackctx.Resolve(trackctx.Inputs{
		TrackArg:   trackArg(f.trackArg, args),
		OutputPath: f.output,
	})
	if err != nil {
		return err
	}

	jPath := f.judgmentsPath
	if jPath == "" {
		strat := f.strategy
		if strat == "" {
			strat = string(judgment.StrategyLexical)
		}
		jPath = tr.JudgmentsPath(strat)
	}

	outPath := f.output
	if outPath == "" {
		strat := f.strategy
		if strat == "" {
			strat = string(judgment.StrategyLexical)
		}
		outPath = tr.QrelsPath(strat)
	}

	jf, err := judgment.ReadFile(jPath)
	if err != nil {
		return fmt.Errorf("read judgments %s: %w", jPath, err)
	}
	if err := judgment.WriteQrels(jf, outPath); err != nil {
		return fmt.Errorf("write qrels: %w", err)
	}
	cmd.Printf("Qrels written: %s\n", outPath)
	return nil
}
