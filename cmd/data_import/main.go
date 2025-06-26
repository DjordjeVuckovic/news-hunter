package main

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/ingest"
	"github.com/DjordjeVuckovic/news-hunter/internal/reader"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"log/slog"
	"os"
)

func main() {
	appSettings := AppSettings{
		ENV: os.Getenv("ENV"),
	}
	cfg, err := appSettings.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	file, err := os.Open(cfg.DataMappingPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		os.Exit(1)
	}

	loader := reader.NewYAMLConfigLoader(file)

	dataFile, err := os.Open(cfg.DatasetPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		os.Exit(1)
	}
	articleReader := reader.NewCSVReader(dataFile)

	mappingCfg, err := loader.Load(true)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	mapper := reader.NewArticleMapper(mappingCfg)

	c := collector.NewArticleCollector(articleReader, mapper)

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

func newPipeline(ctx context.Context, cfg *DataImportConfig, coll collector.Collector[domain.Article]) (ingest.Pipeline, error) {
	switch cfg.StorageType {
	case storage.PG:
		pg, err := storage.NewPgStorer(ctx, *cfg.Postgres)
		if err != nil {
			return nil, err
		}
		return ingest.NewPgPipeline(coll, pg, ingest.WithPgBulk(cfg.BulkOptions.Size)), nil

	case storage.ES:
		es, err := storage.NewEsStorer(*cfg.Elasticsearch)
		if err != nil {
			return nil, err
		}
		return ingest.NewEsPipeline(coll, es, ingest.WithESBulk(cfg.BulkOptions.Size)), nil
	case storage.InMem:
		return nil, storage.ErrUnsupportedStorer
	default:
		return nil, storage.ErrUnsupportedStorer
	}
}
