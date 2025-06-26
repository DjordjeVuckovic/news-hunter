package config

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type StorageType string

const (
	StorageTypeEs    StorageType = "es"
	StorageTypePg    StorageType = "pg"
	StorageTypeInMem             = "in_mem"
)

type DataImportConfig struct {
	StorageType   StorageType
	Elasticsearch *struct {
		AddressesCommaSeparated string
		Username                string
		Password                string
		IndexName               string
	}
	Postgres *struct {
		ConnectionString string
	}
	DatasetPath     string
	DataMappingPath string
	BulkOptions     *struct {
		Enabled bool
		Size    int
	}
}

func LoadConfig() (*DataImportConfig, error) {
	err := godotenv.Load("cmd/data_import/.env")
	if err != nil {
		if os.Getenv("ENV") == "local" {
			slog.Error("Failed to load environment variables in production mode", "error", err)
			return nil, err
		}
		slog.Debug("Skipping .env ...", "error", err)
	}

	storageType := (StorageType)(os.Getenv("STORAGE_TYPE"))
	// Load configuration from environment variables

	return &DataImportConfig{
		StorageType: storageType,
		Elasticsearch: &struct {
			AddressesCommaSeparated string
			Username                string
			Password                string
			IndexName               string
		}{
			AddressesCommaSeparated: "http://localhost:9200",
			Username:                "elastic",
			Password:                "password",
			IndexName:               "news_articles",
		},
		DatasetPath:     "/path/to/dataset.json",
		DataMappingPath: "/path/to/data_mapping.json",
		BulkOptions: &struct {
			Enabled bool
			Size    int
		}{
			Enabled: true,
			Size:    1000,
		},
	}, nil
}
