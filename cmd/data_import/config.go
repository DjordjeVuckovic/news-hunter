package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage/factory"
	"github.com/DjordjeVuckovic/news-hunter/pkg/config/env"
)

func NewAppConfig() *AppConfig {
	return &AppConfig{
		ENV: os.Getenv("ENV"),
	}
}

type AppConfig struct {
	ENV string
}

type DataImportConfig struct {
	DatasetPath     string
	DataMappingPath string
	BulkOptions     *struct {
		Enabled bool
		Size    int
	}
	factory.StorageConfig
}

func (as *AppConfig) Load() (*DataImportConfig, error) {
	err := env.LoadDotEnv(as.ENV, "cmd/data_import/.env", "cmd/data_import/pg.env", "cmd/data_import/rs.env")

	if err != nil {
		slog.Info("Skipping .env environment variables...", "error", err)
	}

	storageCfg, err := factory.LoadEnv()
	if err != nil {
		slog.Error("Failed to load storage configuration from environment", "error", err)
		return nil, err
	}

	mappingPath := os.Getenv("MAPPING_CONFIG_PATH")
	if mappingPath == "" {
		slog.Error("MAPPING_CONFIG_PATH environment variable is not set")
		return nil, fmt.Errorf("MAPPING_CONFIG_PATH environment variable is not set")
	}

	dsPath := os.Getenv("DATASET_PATH")
	if dsPath == "" {
		slog.Error("DATASET_PATH environment variable is not set")
		return nil, fmt.Errorf("DATASET_PATH environment variable is not set")
	}

	bulkEnabled := os.Getenv("BULK_ENABLED")
	bulkSize := os.Getenv("BULK_SIZE")
	bulkSizeNum, err := strconv.Atoi(bulkSize)
	if err != nil {
		bulkSizeNum = 5_000
	}

	cfg := &DataImportConfig{
		DatasetPath:     dsPath,
		DataMappingPath: mappingPath,
		BulkOptions: &struct {
			Enabled bool
			Size    int
		}{
			Enabled: bulkEnabled == "true",
			Size:    bulkSizeNum,
		},
		StorageConfig: *storageCfg,
	}

	return cfg, nil
}
