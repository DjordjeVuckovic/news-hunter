package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type cliConfig struct {
	SpecPath    string
	SuitePath   string
	PgConnStr   string
	EsAddresses string
	EsIndex     string
	KValues     string
	MaxK        int
	Warmup      int
	Runs        int
	Output      string
	Mode        string
	PoolPath    string
}

func parseFlags() cliConfig {
	cfg := cliConfig{}

	flag.StringVar(&cfg.SpecPath, "spec", "", "Path to bench spec YAML (multi-job mode)")
	flag.StringVar(&cfg.SuitePath, "suite", "configs/bench/fts_quality_v1.yaml", "Path to bench suite YAML (quick single-job mode)")
	flag.StringVar(&cfg.PgConnStr, "pg", "", "PostgreSQL connection string")
	flag.StringVar(&cfg.EsAddresses, "es-addresses", "", "Elasticsearch addresses, comma-separated")
	flag.StringVar(&cfg.EsIndex, "es-index", "news", "Elasticsearch index name")
	flag.StringVar(&cfg.KValues, "k", "3,5,10", "K values for metrics, comma-separated")
	flag.IntVar(&cfg.MaxK, "max-k", 100, "Maximum number of results to retrieve per query")
	flag.IntVar(&cfg.Warmup, "warmup", 0, "Number of warmup runs before measurement")
	flag.IntVar(&cfg.Runs, "runs", 1, "Number of measured iterations (median latency used)")
	flag.StringVar(&cfg.Output, "output", "", "Output path for results (JSON file or pool YAML)")
	flag.StringVar(&cfg.Mode, "mode", "bench", "Run mode: bench, pool, or judge")
	flag.StringVar(&cfg.PoolPath, "pool", "", "Path to pool file (for judge mode)")

	flag.Parse()
	return cfg
}

func (c cliConfig) parseKValues() ([]int, error) {
	parts := strings.Split(c.KValues, ",")
	vals := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid k value %q: %w", p, err)
		}
		if v <= 0 {
			return nil, fmt.Errorf("k value must be positive, got %d", v)
		}
		vals = append(vals, v)
	}
	return vals, nil
}
