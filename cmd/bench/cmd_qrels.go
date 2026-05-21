package main

import (
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/judgment"
	"github.com/spf13/cobra"
)

type qrelsFlags struct {
	judgmentsPath string
	output        string
}

func newQrelsCmd() *cobra.Command {
	var f qrelsFlags
	cmd := &cobra.Command{
		Use:     "qrels",
		Short:   "Export judgments to TREC qrels format",
		Long:    "Converts a judgment YAML to standard TREC qrels (query_id 0 doc_id grade) for use with trec_eval and external IR tooling.",
		Example: "  bench qrels --judgments annotations_v1.yaml --output qrels_v1.tsv",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeQrels(cmd, f)
		},
	}
	cmd.Flags().StringVar(&f.judgmentsPath, "judgments", "", "Path to judgment YAML (required)")
	cmd.Flags().StringVar(&f.output, "output", "", "Output qrels TSV path (required)")
	_ = cmd.MarkFlagRequired("judgments")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func executeQrels(cmd *cobra.Command, f qrelsFlags) error {
	jf, err := judgment.ReadFile(f.judgmentsPath)
	if err != nil {
		return fmt.Errorf("read judgments: %w", err)
	}
	if err := judgment.WriteQrels(jf, f.output); err != nil {
		return fmt.Errorf("write qrels: %w", err)
	}
	cmd.Printf("Qrels written: %s\n", f.output)
	return nil
}
