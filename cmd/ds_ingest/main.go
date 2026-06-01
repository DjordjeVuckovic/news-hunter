package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/ingest"
	"github.com/DjordjeVuckovic/news-hunter/internal/ingest/reader"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/document"
)

func main() {
	appSettings := NewAppConfig()

	cfg, err := appSettings.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataFile, err := os.Open(cfg.DatasetPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		os.Exit(1)
	}
	var articleReader reader.RawParallelReader
	switch filepath.Ext(cfg.DatasetPath) {
	case ".jsonl":
		articleReader = reader.NewJSONLReader(dataFile)
	default:
		articleReader = reader.NewCSVReader(dataFile)
	}

	mapper, err := newMapper(cfg)
	if err != nil {
		slog.Error("failed to create mapper", "error", err)
		os.Exit(1)
	}

	c := ingest.NewArticleCollector(articleReader, mapper)

	pipeline, err := newPipeline(ctx, cfg, c)
	if err != nil {
		slog.Error("failed to create pipeline", "error", err)
		os.Exit(1)
	}

	e := pipeline.Run(ctx)

	if e != nil {
		slog.Error("failed to run pipeline", "error", e)
		os.Exit(1)
	}

}

// newMapper selects the record-to-Article mapper. When mapping is disabled the
// dataset is assumed to already be canonical (produced by cmd/preprocessor), so
// the direct mapper is used and no YAML config is required.
func newMapper(cfg *DataImportConfig) (reader.Mapper, error) {
	if !cfg.MappingEnabled {
		slog.Info("Mapping disabled — using direct mapper (expects canonical dataset)")
		return reader.NewArticleDirectMapper(), nil
	}

	file, err := os.Open(cfg.DataMappingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open mapping config: %w", err)
	}
	defer file.Close()

	mappingCfg, err := reader.NewYAMLConfigLoader(file).Load(true)
	if err != nil {
		return nil, fmt.Errorf("failed to load mapping config: %w", err)
	}
	return reader.NewArticleMapper(mappingCfg), nil
}

func newPipeline(
	ctx context.Context,
	cfg *DataImportConfig,
	coll ingest.Collector[document.Article]) (ingest.Pipeline, error) {
	slog.Info("Creating pipeline", "storageType", cfg.StorageConfig.Type)

	storer, err := factory.NewIndexer(ctx, cfg.StorageConfig)
	if err != nil {
		slog.Error("failed to create storer", "error", err)
		return nil, err
	}

	var opts []ingest.PipelineOption
	if cfg.BulkOptions.Enabled {
		opts = append(opts, ingest.WithBulk(cfg.BulkOptions.Size))
	}

	if cfg.Embedding.Enabled {
		ollama, err := embedding.NewOllamaClient(cfg.Embedding.BaseURL)
		if err != nil {
			slog.Error("failed to create embedder", "error", err)
			return nil, err
		}
		embedder := embedding.NewEmbedder(ollama)
		storageEmbedder, err := factory.NewEmbedderIndexer(ctx, cfg.StorageConfig)
		if err != nil {
			slog.Error("storer does not support embedding")
			return nil, err
		}
		opts = append(opts, ingest.WithEmbeddings(storageEmbedder, embedder))
	}

	return ingest.NewPipeline(coll, storer, opts...), nil
}
