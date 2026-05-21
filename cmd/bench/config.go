package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
)

// quickSpecFlags collects engine connection info from CLI flags when the user
// hasn't supplied a --spec file. Lets users run a single-job benchmark without
// authoring a full BenchSpec YAML.
type quickSpecFlags struct {
	suitePath   string
	pgConnStr   string
	esAddresses string
	esIndex     string
	apiBaseURL  string
}

func (q quickSpecFlags) build() (*spec.BenchSpec, error) {
	if q.pgConnStr == "" && q.esAddresses == "" && q.apiBaseURL == "" {
		return nil, fmt.Errorf("at least one of --pg, --es-addresses, --api must be set when --spec is not used")
	}
	engines := make(map[string]spec.Engine)
	var names []string

	if q.pgConnStr != "" {
		engines["pg"] = spec.Engine{Type: "postgres", Connection: q.pgConnStr}
		names = append(names, "pg")
	}
	if q.esAddresses != "" {
		idx := q.esIndex
		if idx == "" {
			idx = "articles"
		}
		engines["elasticsearch"] = spec.Engine{Type: "elasticsearch", Connection: q.esAddresses, Index: idx}
		names = append(names, "elasticsearch")
	}
	if q.apiBaseURL != "" {
		engines["api"] = spec.Engine{Type: "api", Connection: q.apiBaseURL}
		names = append(names, "api")
	}
	return &spec.BenchSpec{
		Engines: engines,
		Jobs: []spec.Job{{
			Name:    "quick",
			Suite:   q.suitePath,
			Engines: names,
		}},
	}, nil
}

func loadBenchSpec(specPath string, quick quickSpecFlags) (*spec.BenchSpec, error) {
	if specPath != "" {
		return spec.LoadFromFile(specPath)
	}
	return quick.build()
}

func createExecutors(ctx context.Context, bs *spec.BenchSpec) (map[string]engine.Executor, func(), error) {
	return engine.CreateFromSpec(ctx, bs.Engines)
}

func parseKList(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		var v int
		if _, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &v); err != nil {
			return nil, fmt.Errorf("invalid k value %q: %w", p, err)
		}
		if v <= 0 {
			return nil, fmt.Errorf("k value must be positive, got %d", v)
		}
		out = append(out, v)
	}
	return out, nil
}

func envOrFlag(envKey, flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv(envKey)
}
