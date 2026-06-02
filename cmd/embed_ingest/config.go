package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	"github.com/DjordjeVuckovic/news-hunter/pkg/config/env"
)

const defaultBatchSize = 5_000

func NewAppConfig() *AppConfig {
	return &AppConfig{ENV: os.Getenv("ENV")}
}

type AppConfig struct {
	ENV string
}

type EmbedIngestConfig struct {
	factory.StorageConfig
	Embedding embedding.Config
	BatchSize int
}

func (as *AppConfig) Load() (*EmbedIngestConfig, error) {
	if err := env.LoadDotEnv(as.ENV, "cmd/embed_ingest/.env"); err != nil {
		slog.Info("Skipping .env environment variables...", "error", err)
	}

	storageCfg, err := factory.LoadEnv()
	if err != nil {
		return nil, err
	}

	embedCfg, err := embedding.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	if embedCfg.Source != embedding.SourceFile {
		return nil, fmt.Errorf("embed_ingest requires EMBEDDING_SOURCE=file, got %q", embedCfg.Source)
	}

	store := embedCfg.ObjectStore
	if store.LocalPath == "" && (store.Bucket == "" || store.Key == "") {
		return nil, fmt.Errorf("set EMBEDDING_FILE_PATH or EMBEDDING_S3_BUCKET + EMBEDDING_S3_KEY")
	}

	batchSize := defaultBatchSize
	if v := os.Getenv("EMBEDDING_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			batchSize = n
		}
	}

	return &EmbedIngestConfig{
		StorageConfig: *storageCfg,
		Embedding:     *embedCfg,
		BatchSize:     batchSize,
	}, nil
}
