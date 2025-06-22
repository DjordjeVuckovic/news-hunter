package main

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/ingest"
	"github.com/DjordjeVuckovic/news-hunter/internal/reader"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

func main() {
	err := godotenv.Load("cmd/data_import/.env")
	if err != nil {
		slog.Error("Failed to load environment variables", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from panic", "error", r)
		}
	}()

	configPath := os.Getenv("MAPPING_CONFIG_PATH")
	file, err := os.Open(configPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		return
	}

	loader := reader.NewYAMLConfigLoader(file)

	dsPath := os.Getenv("DATASET_PATH")
	dataFile, err := os.Open(dsPath)
	if err != nil {
		slog.Error("failed to read configuration file", "error", err)
		return
	}
	articleReader := reader.NewCSVReader(dataFile)

	cfg, err := loader.Load(true)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return
	}
	mapper := reader.NewArticleMapper(cfg)

	c := collector.NewArticleCollector(articleReader, mapper)

	connStr := os.Getenv("DB_CONNECTION_STRING")
	db, err := storage.NewPgStorer(ctx, connStr)
	// db := storage.NewJsonFileStorer("")

	pipeline := ingest.NewPgPipeline(c, db, ingest.WithPgBulk(1_000))

	e := pipeline.Run(ctx)

	if e != nil {
		slog.Error("failed to run pipeline", "error", e)
		return
	}

}
