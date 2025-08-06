package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
	"github.com/DjordjeVuckovic/news-hunter/pkg/config/env"
)

type AppConfig struct {
	ENV string
}

func NewAppSettings() *AppConfig {
	return &AppConfig{
		ENV: os.Getenv("ENV"),
	}
}

type NewsSearchConfig struct {
	StorageType   storage.Type
	Elasticsearch *es.Config
	Postgres      *pg.Config
}

func (as *AppConfig) Load() (*NewsSearchConfig, error) {

	err := env.LoadDotEnv(as.ENV, "cmd/news_search/.env")

	storageType := (storage.Type)(os.Getenv("STORAGE_TYPE"))
	if storageType == "" {
		slog.Error("STORAGE_TYPE environment variable is not set")
		return nil, err
	}
	if storageType != storage.ES && storageType != storage.PG && storageType != storage.InMem {
		slog.Error("Invalid STORAGE_TYPE environment variable value", "value", storageType, "expected", []storage.Type{storage.ES, storage.PG, storage.InMem})
		return nil, err
	}
	var esCfg *es.Config
	if storageType == storage.ES {
		esCfg = &es.Config{
			Addresses: strings.Split(os.Getenv("ES_ADDRESSES"), ","),
			IndexName: os.Getenv("ES_INDEX_NAME"),
			Username:  os.Getenv("ES_USERNAME"),
			Password:  os.Getenv("ES_PASSWORD"),
		}
		if len(esCfg.Addresses) == 0 || esCfg.IndexName == "" {
			slog.Error("Elasticsearch configuration is incomplete", "addresses", esCfg.Addresses, "indexName", esCfg.IndexName)
			return nil, err
		}
	}

	var pgCfg *pg.Config
	if storageType == storage.PG {
		pgCfg = &pg.Config{
			ConnStr: os.Getenv("PG_CONNECTION_STRING"),
		}
		if pgCfg.ConnStr == "" {
			slog.Error("PostgreSQL connection string is not set")
			return nil, err
		}
	}

	return &NewsSearchConfig{
		StorageType:   storageType,
		Postgres:      pgCfg,
		Elasticsearch: esCfg,
	}, nil
}
