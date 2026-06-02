package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
)

func createExecutors(ctx context.Context, bs *spec.BenchSpec) (map[string]engine.Executor, func(), error) {
	return engine.CreateFromSpec(ctx, bs.Engines)
}

// buildQueryVectorStore builds the vector store pool/run use to embed queries
// for vector/hybrid tracks (PG precedence: it borrows a postgres engine's
// connection). Returns (nil, nil) when EMBEDDING_BASE_URL is unset or the spec
// has no postgres engine — tracks without vector queries don't need it, and
// vector queries without it simply fail to resolve (logged per-engine).
func buildQueryVectorStore(ctx context.Context, bs *spec.BenchSpec) (storage.VectorStore, error) {
	baseURL := os.Getenv("EMBEDDING_BASE_URL")
	if baseURL == "" {
		return nil, nil
	}
	var pgConn string
	for _, eng := range bs.Engines {
		if eng.Type == "postgres" {
			pgConn = eng.Connection
			break
		}
	}
	if pgConn == "" {
		return nil, nil
	}
	client, err := embedding.NewOllamaClient(baseURL)
	if err != nil {
		return nil, fmt.Errorf("embedding client: %w", err)
	}
	return factory.NewVectorStore(ctx, factory.VectorStoreConfig{
		PgConnStr:       pgConn,
		EmbeddingClient: client,
		Model:           os.Getenv("EMBEDDING_MODEL"),
	})
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
