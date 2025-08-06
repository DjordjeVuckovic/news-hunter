package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
	"github.com/DjordjeVuckovic/news-hunter/pkg/config/env"
)

func NewAppSettings() *AppConfig {
	return &AppConfig{
		ENV: os.Getenv("ENV"),
	}
}

type AppConfig struct {
	ENV string
}

type DataImportConfig struct {
	StorageType     storage.Type
	Elasticsearch   *es.Config
	Postgres        *pg.Config
	DatasetPath     string
	DataMappingPath string
	BulkOptions     *struct {
		Enabled bool
		Size    int
	}
}

func (as *AppConfig) Load() (*DataImportConfig, error) {
	err := env.LoadDotEnv(as.ENV, "cmd/data_import/.env")
	if err != nil {
		slog.Error("Failed to .env load environment variables", "error", err)
		return nil, err
	}

	storageType := (storage.Type)(os.Getenv("STORAGE_TYPE"))
	if storageType == "" {
		slog.Error("STORAGE_TYPE environment variable is not set")
		return nil, err
	}
	if storageType != storage.ES && storageType != storage.PG && storageType != storage.InMem {
		slog.Error("Invalid STORAGE_TYPE environment variable value", "value", storageType, "expected", []storage.Type{storage.ES, storage.PG, storage.InMem})
		return nil, err
	}

	mappingPath := os.Getenv("MAPPING_CONFIG_PATH")
	if mappingPath == "" {
		slog.Error("MAPPING_CONFIG_PATH environment variable is not set")
		return nil, err
	}

	dsPath := os.Getenv("DATASET_PATH")
	if dsPath == "" {
		slog.Error("DATASET_PATH environment variable is not set")
		return nil, err
	}

	bulkEnabled := os.Getenv("BULK_ENABLED")
	bulkSize := os.Getenv("BULK_SIZE")
	bulkSizeNum, err := strconv.Atoi(bulkSize)
	if err != nil {
		bulkSizeNum = 5000 // Default bulk size if not set or invalid
	}

	cfg := &DataImportConfig{
		StorageType:     storageType,
		DatasetPath:     dsPath,
		DataMappingPath: mappingPath,
		BulkOptions: &struct {
			Enabled bool
			Size    int
		}{
			Enabled: bulkEnabled == "true",
			Size:    bulkSizeNum,
		},
	}

	if storageType == storage.ES {
		cfg.Elasticsearch = &es.Config{
			Addresses: strings.Split(os.Getenv("ES_ADDRESSES"), ","),
			IndexName: os.Getenv("ES_INDEX_NAME"),
			Username:  os.Getenv("ES_USERNAME"),
			Password:  os.Getenv("ES_PASSWORD"),
		}
		if len(cfg.Elasticsearch.Addresses) == 0 || cfg.Elasticsearch.IndexName == "" {
			slog.Error("Elasticsearch configuration is incomplete", "addresses", cfg.Elasticsearch.Addresses, "indexName", cfg.Elasticsearch.IndexName)
			return nil, err
		}
	}

	if storageType == storage.PG {
		cfg.Postgres = &pg.Config{
			ConnStr: os.Getenv("PG_CONNECTION_STRING"),
		}
		if cfg.Postgres.ConnStr == "" {
			slog.Error("PostgreSQL connection string is not set")
			return nil, err
		}
	}

	return cfg, nil
}
