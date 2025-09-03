package main

import (
	"log/slog"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	"github.com/DjordjeVuckovic/news-hunter/pkg/config/env"
)

type AppConfig struct {
	ENV string
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		ENV: os.Getenv("ENV"),
	}
}

type NewsSearchConfig struct {
	StorageConfig factory.StorageConfig
}

func (as *AppConfig) Load() (*NewsSearchConfig, error) {

	err := env.LoadDotEnv(as.ENV, "cmd/news_search/.env")
	if err != nil {
		slog.Info("Failed to .env load environment variables, continuing with existing environment variables", "error", err)
	}

	storageCfg, err := factory.LoadFromEnv()
	if err != nil {
		slog.Error("Failed to load storage configuration from environment", "error", err)
		return nil, err
	}

	return &NewsSearchConfig{
		StorageConfig: *storageCfg,
	}, nil
}
