package main

import (
	"context"
	"log/slog"
	"os"

	collector2 "github.com/DjordjeVuckovic/news-hunter/internal/ingest/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/ingest/processor"
	reader2 "github.com/DjordjeVuckovic/news-hunter/internal/ingest/reader"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/document"
)

func main() {
	appSettings := NewAppConfig()

	cfg, err := appSettings.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file, err := os.Open(cfg.DataMappingPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		os.Exit(1)
	}

	loader := reader2.NewYAMLConfigLoader(file)

	dataFile, err := os.Open(cfg.DatasetPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		os.Exit(1)
	}
	articleReader := reader2.NewCSVReader(dataFile)

	mappingCfg, err := loader.Load(true)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	mapper := reader2.NewArticleMapper(mappingCfg)

	c := collector2.NewArticleCollector(articleReader, mapper)

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

func newPipeline(
	ctx context.Context,
	cfg *DataImportConfig,
	coll collector2.Collector[document.Article]) (processor.Pipeline, error) {
	slog.Info("Creating pipeline", "storageType", cfg.StorageConfig.Type)

	storer, err := factory.NewIndexer(ctx, cfg.StorageConfig)
	if err != nil {
		slog.Error("failed to create storer", "error", err)
		return nil, err
	}

	return processor.NewPipeline(coll, storer, processor.WithBulk(cfg.BulkOptions.Size)), nil
}
