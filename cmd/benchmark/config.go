package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type cliConfig struct {
	SuitePath   string
	PgConnStr   string
	EsAddresses string
	EsIndex     string
	KValues     string
	MaxK        int
}

func parseFlags() cliConfig {
	cfg := cliConfig{}

	flag.StringVar(&cfg.SuitePath, "suite", "configs/benchmark/fts_quality_v1.yaml", "Path to benchmark suite YAML")
	flag.StringVar(&cfg.PgConnStr, "pg", "", "PostgreSQL connection string (enables PG native engine)")
	flag.StringVar(&cfg.EsAddresses, "es-addresses", "", "Elasticsearch addresses, comma-separated (enables ES engine)")
	flag.StringVar(&cfg.EsIndex, "es-index", "news", "Elasticsearch index name")
	flag.StringVar(&cfg.KValues, "k", "3,5,10", "K values for metrics, comma-separated")
	flag.IntVar(&cfg.MaxK, "max-k", 10, "Maximum number of results to retrieve per query")

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
