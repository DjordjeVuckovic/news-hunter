package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
)

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

// trackArg picks up the track from a flag or first positional arg. The CLI
// allows either form: `bench pool fts_quality` or `bench pool --track fts_quality`.
func trackArg(flag string, args []string) string {
	if flag != "" {
		return flag
	}
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
